package raft

type Message interface{}

type Command interface{}

type Stop struct{}

type LeadershipTimeout struct{}

type ElectionTimeout struct{}

type RequestVoteResult struct {
	nodeID NodeID
	term   Term
	accept bool
}

type RequestVote struct {
	term         Term
	candidateID  NodeID
	lastLogIndex int
	lastLogTerm  Term
}

type AppendEntries struct {
	term         Term
	leaderID     NodeID
	prevLogIndex int
	prevLogTerm  Term
	entries      []LogEntry
	leaderCommit int
}

type AppendEntriesResult struct {
	nodeID  NodeID
	term    Term
	success bool
}
