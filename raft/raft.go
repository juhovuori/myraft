package raft

type State string
type Term int
type LogEntry struct {
	Term    Term
	Command string
}

type Command interface{}

type Stop struct{}

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
	node.electionTimer = NewDefaultTimer(ElectionMinTimeout, ElectionMaxTimeout, TimerCallback(node.OnElectionTimeout), &node)
	node.leadershipTimer = NewDefaultTimer(LeadershipMinTimeout, LeadershipMaxTimeout, TimerCallback(node.OnLeadershipTimeout), &node)
	return &node
}

func (n *SimpleRaftNode) Stop() {
	n.commands <- Stop{}
	<-n.done
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
		response := n.getVoteRequestResponse(cmd)
		n.comm.Send(cmd.candidateID, RequestVoteResult{n.currentTerm, response})
	case RequestVoteResult:
		n.Log("Got vote", cmd)
		n.receivedVotes++
		if n.receivedVotes >= n.nodeCount/2+1 {
			n.ChangeState(Leader)
		}
	case AppendEntries:
		n.Log("AppendEntries", cmd)
		n.electionTimer.Reset()
	case AppendEntriesResult:
		n.Log("AppendEntriesResult", cmd)
	case Stop:
		return false
	default:
		n.Log("Invalid command", cmd)
	}
	return true
}

func (n *SimpleRaftNode) OnLeadershipTimeout() {
	switch n.state {
	case Leader:
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
	default:
		n.Log("Leader timeout => ignoring, state is", n.state)
	}
}

func (n *SimpleRaftNode) OnElectionTimeout() {
	switch n.state {
	case Follower:
		n.Log("Election timeout => becoming candidate")
		n.ChangeState(Candidate)
	case Candidate:
		n.Log("Election timeout => restart election")
		n.ChangeState(Candidate)
	default:
		panic("Election timeout => invalid state")
	}
}

func (n *SimpleRaftNode) ChangeState(newState State) {
	n.Logf("Becoming %v (was %v)", newState, n.state)
	switch newState {
	case Follower:
		n.state = Follower
	case Candidate:
		n.state = Candidate
		n.currentTerm++
		n.vote(n.ID())
		n.electionTimer.Reset()
		n.receivedVotes = 0
		msg := RequestVote{n.currentTerm, n.ID(), 0, n.currentTerm}
		n.comm.Broadcast(msg)
	case Leader:
		n.state = Leader
		n.OnLeadershipTimeout()
	}
}

func (n *SimpleRaftNode) OnMessage(m Message) {
	n.commands <- m
}

func (n *SimpleRaftNode) vote(nodeID NodeID) {
	n.votedFor = &nodeID
}

func (n *SimpleRaftNode) getVoteRequestResponse(r RequestVote) bool {
	n.Log("Got vote request", r)
	if r.term < n.currentTerm {
		return false
	}
	return (n.votedFor == nil || *n.votedFor == r.candidateID) && (r.lastLogIndex >= len(n.log)-1 && r.lastLogTerm >= n.currentTerm)
	// TODO: double check log up to date comparison
}
