package raft

import "github.com/juhovuori/myraft/comm"

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
	comm.Node
	Start()
	Stop()
}

type SimpleRaftNode struct {
	comm.SimpleNode
	state           State
	currentTerm     Term
	votedFor        *comm.NodeID
	log             []LogEntry
	electionTimer   Timer
	leadershipTimer Timer
	comm            comm.Comm
	nodeCount       int
	receivedVotes   int
	commitIndex     int
	lastApplied     int
	done            chan bool
}

func NewSimpleRaftNode(nodeID comm.NodeID, c comm.Comm, nodeCount int) *SimpleRaftNode {
	node := SimpleRaftNode{
		*comm.NewSimpleNode(nodeID),
		Follower,
		Term(0),
		nil,
		[]LogEntry{},
		nil,
		nil,
		c,
		nodeCount,
		0,
		0,
		0,
		make(chan bool),
	}
	node.electionTimer = NewDefaultTimer("electiontimer", ElectionMinTimeout, ElectionMaxTimeout, TimerCallback(node.OnElectionTimeout), nil)
	node.leadershipTimer = NewDefaultTimer("leadershiptimer", LeadershipMinTimeout, LeadershipMaxTimeout, TimerCallback(node.OnLeadershipTimeout), nil)
	return &node
}

func (n *SimpleRaftNode) OnLeadershipTimeout() {
	//TODO: n.commands <- LeadershipTimeout{}
}

func (n *SimpleRaftNode) OnElectionTimeout() {
	//TODO:n.commands <- ElectionTimeout{}
}

func (n *SimpleRaftNode) OnRPC(m interface{}) interface{} {
	switch m := m.(type) {
	case AppendEntries:
		return n.OnAppendEntries(m)
	case RequestVote:
		return n.OnRequestVote(m)
	default:
		return InvalidRPC{}
	}
}

func (n *SimpleRaftNode) OnRequestVote(rv RequestVote) RequestVoteResult {
	n.Logf("Got vote request %+v", rv)
	var response bool
	if rv.term < n.currentTerm {
		// Candidate term out of date
		n.Log("Candidate term out of date")
		response = false
	} else if n.votedFor != nil && *n.votedFor != rv.candidateID {
		// Voted for someone else
		n.Log("Voted for someone else")
		response = false
	} else if rv.lastLogTerm < n.currentTerm || (rv.lastLogTerm == n.currentTerm && rv.lastLogIndex == len(n.log)-1) {
		// Candidate log not as up to date as mine
		n.Log("Candidate log not as up to date as mine", rv.lastLogTerm, n.currentTerm)
		response = false
	} else {
		response = true
	}
	n.assertUpToDateTerm(rv.term)
	n.vote(rv.candidateID)
	return RequestVoteResult{n.ID(), n.currentTerm, response}
	/*	case RequestVoteResult:
		n.Logf("Got response to request vote %+v", rv)
		n.assertUpToDateTerm(rv.term)
		if rv.accept {
			n.assertUpToDateTerm(rv.term)
			n.receivedVotes++
			if n.receivedVotes >= (n.nodeCount/2)+1 {
				n.changeState(Leader)
			}
		}
	*/
}

func (n *SimpleRaftNode) OnAppendEntries(ae AppendEntries) AppendEntriesResult {
	n.Logf("AppendEntries %+v", ae)
	n.assertUpToDateTerm(ae.term)
	n.electionTimer.Reset()
	//TODO: Send result properly
	return AppendEntriesResult{}
	/*
		case AppendEntriesResult:
			n.Logf("AppendEntriesResult %+v", ae)
			n.assertUpToDateTerm(ae.term)
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
			n.Logf("Invalid command %+v", ae)
		}
		return true
	*/
}

func (n *SimpleRaftNode) Stop() {
	n.done <- true
}

func (n *SimpleRaftNode) Start() {
	n.electionTimer.Reset()
	n.Log("Started raft")
	<-n.done
	n.Log("Stopped raft")
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
		n.comm.BroadcastRPC(msg)
	case Leader:
		n.state = Leader
		// TODO: n.RunCommand(LeadershipTimeout{})
	}
}

func (n *SimpleRaftNode) vote(nodeID comm.NodeID) {
	n.votedFor = &nodeID
}

func (n *SimpleRaftNode) assertUpToDateTerm(term Term) {
	if term > n.currentTerm {
		n.Log("interface{} with term > current term => adjusting")
		n.currentTerm = term
		n.changeState(Follower)
	}
}
