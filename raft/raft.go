package raft

type State string
type Term int
type LogEntry struct {
	Term    Term
	Command string
}

type Command interface{}

type Stop struct{}

type LeadershipTimeout struct{}

type ElectionTimeout struct{}

const (
	Follower  = State("follower")
	Candidate = State("candidate")
	Leader    = State("leader")
)

type RaftNode interface {
	Node
	Raft(comm Comm, cluster Cluster)
	Stop()
}

type SimpleRaftNode struct {
	SimpleNode
	state           State
	currentTerm     Term
	votedFor        *NodeID
	log             []LogEntry
	electionTimer   Timer
	leadershipTimer Timer
	comm            Comm
	nodeCount       int
	receivedVotes   int
	commitIndex     int
	lastApplied     int
	commands        chan Command
	done            chan bool
}

func NewSimpleRaftNode(nodeID NodeID, comm Comm, nodeCount int) *SimpleRaftNode {
	node := SimpleRaftNode{
		*NewSimpleNode(nodeID),
		Follower,
		Term(0),
		nil,
		[]LogEntry{},
		nil,
		nil,
		comm,
		nodeCount,
		0,
		0,
		0,
		make(chan Command),
		make(chan bool),
	}
	node.electionTimer = NewDefaultTimer("electiontimer", ElectionMinTimeout, ElectionMaxTimeout, TimerCallback(node.OnElectionTimeout), &node)
	node.leadershipTimer = NewDefaultTimer("leadershiptimer", LeadershipMinTimeout, LeadershipMaxTimeout, TimerCallback(node.OnLeadershipTimeout), &node)
	return &node
}

func (n *SimpleRaftNode) OnLeadershipTimeout() {
	n.commands <- LeadershipTimeout{}
}

func (n *SimpleRaftNode) OnElectionTimeout() {
	n.commands <- ElectionTimeout{}
}

func (n *SimpleRaftNode) OnMessage(m Message) {
	n.commands <- m
}

func (n *SimpleRaftNode) Stop() {
	n.Log("stop 11111")
	n.commands <- Stop{}
	n.Log("stop 00000")
	<-n.done
	n.Log("stop 99999")
}

func (n *SimpleRaftNode) Raft(comm Comm, cluster Cluster) {
	n.electionTimer.Reset()
	n.Log("Started raft")
	for cmd := range n.commands {
		if !n.RunCommand(cmd) {
			break
		}
	}
	n.Log("Stopped raft")
	n.done <- true
}

func (n *SimpleRaftNode) RunCommand(cmd Command) bool {
	switch cmd := cmd.(type) {
	case RequestVote:
		n.Log("Got vote request", cmd)
		var response bool
		if cmd.term < n.currentTerm {
			response = false
		} else {
			response = (n.votedFor == nil || *n.votedFor == cmd.candidateID) &&
				cmd.lastLogIndex >= len(n.log)-1 &&
				cmd.lastLogTerm >= n.currentTerm
			// TODO: double check log up to date comparison
		}
		n.comm.Send(cmd.candidateID, RequestVoteResult{n.currentTerm, response})
	case RequestVoteResult:
		n.Log("Got vote", cmd)
		n.receivedVotes++
		if n.receivedVotes >= (n.nodeCount/2)+1 {
			n.changeState(Leader)
		}
	case AppendEntries:
		n.Log("AppendEntries", cmd)
		n.electionTimer.Reset()
	case AppendEntriesResult:
		n.Log("AppendEntriesResult", cmd)
	case Stop:
		n.Log("stop")
		return false
	case LeadershipTimeout:
		n.Log("Leader timeout", n.state)
		if n.state != Leader {
			n.Log("Ignoring, state is", n.state)
			break
		}
		prevLogTerm := Term(-1)
		if len(n.log) > 0 {
			prevLogTerm = n.log[len(n.log)-1].Term
		}
		msg := AppendEntries{
			term:         n.currentTerm,
			leaderID:     n.ID(),
			prevLogIndex: len(n.log) - 1,
			prevLogTerm:  prevLogTerm,
			entries:      []LogEntry{},
			leaderCommit: n.commitIndex,
		}
		n.comm.Broadcast(msg)
		n.leadershipTimer.Reset()
	case ElectionTimeout:
		n.Log("Election timeout")
		n.changeState(Candidate) // restart election
	default:
		n.Log("Invalid command", cmd)
	}
	return true
}

func (n *SimpleRaftNode) changeState(newState State) {
	n.Logf("Becoming %v (was %v)", newState, n.state)
	switch newState {
	case Follower:
		n.electionTimer.Stop() // race condition, what if timeout command was already queued
		n.state = Follower
	case Candidate:
		n.electionTimer.Reset()
		n.state = Candidate
		n.currentTerm++
		nodeID := n.ID()
		n.votedFor = &nodeID
		n.receivedVotes = 0
		msg := RequestVote{n.currentTerm, n.ID(), 0, n.currentTerm}
		n.comm.Broadcast(msg)
	case Leader:
		n.electionTimer.Stop() // race condition, what if timeout command was already queued
		n.state = Leader
		n.RunCommand(LeadershipTimeout{})
	}
}
