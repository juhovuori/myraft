package raft

type State string
type Term int
type LogEntry struct {
	Term    Term
	Command string
}

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
	node.electionTimer = NewDefaultTimer("electiontimer", ElectionMinTimeout, ElectionMaxTimeout, TimerCallback(node.OnElectionTimeout), nil)
	node.leadershipTimer = NewDefaultTimer("leadershiptimer", LeadershipMinTimeout, LeadershipMaxTimeout, TimerCallback(node.OnLeadershipTimeout), nil)
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
		n.Logf("Got vote request %+v", cmd)
		var response bool
		if cmd.term < n.currentTerm {
			// Candidate term out of date
			n.Log("Candidate term out of date")
			response = false
		} else if n.votedFor != nil && *n.votedFor != cmd.candidateID {
			// Voted for someone else
			n.Log("Voted for someone else")
			response = false
		} else if cmd.lastLogTerm < n.currentTerm || (cmd.lastLogTerm == n.currentTerm && cmd.lastLogIndex == len(n.log)-1) {
			// Candidate log not as up to date as mine
			n.Log("Candidate log not as up to date as mine", cmd.lastLogTerm, n.currentTerm)
			response = false
		} else {
			response = true
		}
		n.assertUpToDateTerm(cmd.term)
		n.vote(cmd.candidateID)
		n.comm.Send(cmd.candidateID, RequestVoteResult{n.ID(), n.currentTerm, response})
	case RequestVoteResult:
		n.Logf("Got response to request vote %+v", cmd)
		n.assertUpToDateTerm(cmd.term)
		if cmd.accept {
			n.assertUpToDateTerm(cmd.term)
			n.receivedVotes++
			if n.receivedVotes >= (n.nodeCount/2)+1 {
				n.changeState(Leader)
			}
		}
	case AppendEntries:
		n.Logf("AppendEntries %+v", cmd)
		n.assertUpToDateTerm(cmd.term)
		n.electionTimer.Reset()
	case AppendEntriesResult:
		n.Logf("AppendEntriesResult %+v", cmd)
		n.assertUpToDateTerm(cmd.term)
	case Stop:
		return false
	case LeadershipTimeout:
		n.Logf("Leader timeout %+v", n.state)
		if n.state != Leader {
			n.Logf("Ignoring, state is %+v", n.state)
			break
		}
		prevLogTerm := Term(0)
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
		if n.votedFor != nil {
			n.Logf("Ignoring, already voted for %+v", *n.votedFor)
		}
		n.changeState(Candidate) // restart election
	default:
		n.Logf("Invalid command %+v", cmd)
	}
	return true
}

func (n *SimpleRaftNode) changeState(newState State) {
	n.Logf("Becoming %v (was %v)", newState, n.state)
	n.electionTimer.Stop() // TODO:race condition, what if timeout command was already queued
	switch newState {
	case Follower:
		n.state = Follower
	case Candidate:
		n.electionTimer.Reset()
		n.state = Candidate
		n.currentTerm++
		n.vote(n.ID())
		n.receivedVotes = 0
		msg := RequestVote{n.currentTerm, n.ID(), 0, n.currentTerm}
		n.Logf("Send vote request %+v", msg)
		n.comm.Broadcast(msg)
	case Leader:
		n.state = Leader
		n.RunCommand(LeadershipTimeout{})
	}
}

func (n *SimpleRaftNode) vote(nodeID NodeID) {
	n.votedFor = &nodeID
}

func (n *SimpleRaftNode) assertUpToDateTerm(term Term) {
	if term > n.currentTerm {
		n.Log("Message with term > current term => adjusting")
		n.currentTerm = term
		n.changeState(Follower)
	}
}
