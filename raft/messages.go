package raft

import "github.com/juhovuori/myraft/comm"

type Command string
type CommandResult struct {
	Success bool
	Leader  comm.NodeID
}
type LeadershipTimeout struct{}

type ElectionTimeout struct{}

type RequestVoteResult struct {
	nodeID comm.NodeID
	term   Term
	accept bool
}

type InvalidRPC struct{}
type RPCTimeout struct{}

type RequestVote struct {
	term         Term
	candidateID  comm.NodeID
	lastLogIndex int
	lastLogTerm  Term
}

type AppendEntries struct {
	term         Term
	leaderID     comm.NodeID
	prevLogIndex int
	prevLogTerm  Term
	entries      []LogEntry
	leaderCommit int
}

type AppendEntriesResult struct {
	nodeID  comm.NodeID
	term    Term
	success bool
}
