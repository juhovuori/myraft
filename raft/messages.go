package raft

type RequestVoteResult struct {
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
	term    Term
	success bool
}
