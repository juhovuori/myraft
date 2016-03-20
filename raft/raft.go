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

type RaftConfig struct {
	Nodes []comm.NodeID
}

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
	config          RaftConfig
	log             []LogEntry
	electionTimer   Timer
	leadershipTimer Timer
	comm            comm.Comm
	commitIndex     int
	lastApplied     int
	neighbours      []comm.NodeID
	done            chan bool
}

func NewSimpleRaftNode(nodeID comm.NodeID, c comm.Comm, config RaftConfig) *SimpleRaftNode {
	neighbours := []comm.NodeID{}
	for _, ID := range config.Nodes {
		if ID != nodeID {
			neighbours = append(neighbours, ID)
		}
	}
	node := SimpleRaftNode{
		*comm.NewSimpleNode(nodeID),
		Follower,
		Term(0),
		nil,
		config,
		[]LogEntry{},
		nil,
		nil,
		c,
		0,
		0,
		neighbours,
		make(chan bool),
	}
	node.electionTimer = NewDefaultTimer("electiontimer", ElectionMinTimeout, ElectionMaxTimeout, TimerCallback(node.OnElectionTimeout), nil)
	node.leadershipTimer = NewDefaultTimer("leadershiptimer", LeadershipMinTimeout, LeadershipMaxTimeout, TimerCallback(node.OnLeadershipTimeout), nil)
	return &node
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

func (n *SimpleRaftNode) OnLeadershipTimeout() {
	if n.state != Leader {
		n.Logf("Ignoring leader timeout, state is %+v", n.state)
		return
	}
	n.Logf("Leader timeout %+v", n.state)
	n.AppendEntries()
}

func (n *SimpleRaftNode) AppendEntries() {
	n.leadershipTimer.Reset()
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
	n.Log("Send append entries")

	responses := n.comm.MulticastRPC(msg, n.neighbours...)
	for count := len(n.neighbours); count > 0; count-- {
		response := <-responses
		payload, ok := response.Payload.(AppendEntriesResult)
		if !ok {
			n.Log("Invalid append entries (possibly timeout)")
			continue
		}
		n.Logf("Got AppendEntriesResult %+v", payload)
		n.assertUpToDateTerm(payload.term)
	}
}

func (n *SimpleRaftNode) OnElectionTimeout() {
	n.Log("Election timeout")
	if n.votedFor != nil {
		n.Logf("Ignoring, already voted for %+v", *n.votedFor)
	}
	n.becomeCandidate() // restart election
}

func (n *SimpleRaftNode) becomeFollower() {
	n.Logf("Becoming follower (was %v)", n.state)
	n.electionTimer.Stop()
	n.state = Follower
}

func (n *SimpleRaftNode) becomeCandidate() {
	n.Logf("Becoming candidate (was %v)", n.state)
	n.electionTimer.Reset()
	n.state = Candidate

	n.currentTerm++
	n.vote(n.ID())
	receivedVotes := 1 // 1 vote from self
	msg := RequestVote{n.currentTerm, n.ID(), 0, n.currentTerm}
	n.Logf("Send vote request %+v", msg)
	count := len(n.neighbours)
	quorum := count/2 + 1
	responses := n.comm.MulticastRPC(msg, n.neighbours...)
	for ; count > 0; count-- {
		response := <-responses
		payload, ok := response.Payload.(RequestVoteResult)
		if !ok {
			n.Log("Invalid vote request response (possibly timeout)")
			continue
		}
		if payload.accept {
			receivedVotes++
			n.Logf("Got yes (%d)", receivedVotes)
		} else {
			n.Log("Got no")
		}
		n.assertUpToDateTerm(payload.term)
		if receivedVotes >= quorum {
			n.becomeLeader()
			return // Rest of the votes can be discarded
		}
	}
}

func (n *SimpleRaftNode) becomeLeader() {
	n.Logf("Becoming leader (was %v)", n.state)
	n.electionTimer.Stop() // TODO:race condition, what if timeout command was already queued
	n.state = Leader
	n.AppendEntries()
}

func (n *SimpleRaftNode) OnRequestVote(rv RequestVote) RequestVoteResult {
	n.Logf("Got vote request %+v", rv)
	n.assertUpToDateTerm(rv.term)
	accept := true
	if rv.term < n.currentTerm {
		// Candidate term out of date
		n.Log("Candidate term out of date")
		accept = false
	} else if n.votedFor != nil && *n.votedFor != rv.candidateID {
		// Voted for someone else
		n.Log("Voted for someone else")
		accept = false
	} else if rv.lastLogTerm < n.currentTerm || (rv.lastLogTerm == n.currentTerm && rv.lastLogIndex == len(n.log)-1) {
		// Candidate log not as up to date as mine
		n.Log("Candidate log not as up to date as mine", rv.lastLogTerm, n.currentTerm)
		accept = false
	} else {
		n.vote(rv.candidateID)
	}
	return RequestVoteResult{n.ID(), n.currentTerm, accept}
}

func (n *SimpleRaftNode) OnAppendEntries(ae AppendEntries) AppendEntriesResult {
	n.Logf("AppendEntries %+v", ae)
	n.assertUpToDateTerm(ae.term)
	n.electionTimer.Reset()
	//TODO: Send result properly
	return AppendEntriesResult{}
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

func (n *SimpleRaftNode) vote(nodeID comm.NodeID) {
	n.votedFor = &nodeID
}

func (n *SimpleRaftNode) assertUpToDateTerm(term Term) {
	if term > n.currentTerm {
		n.Log("Message with term > current term => adjusting")
		n.currentTerm = term
		n.becomeFollower()
	}
}
