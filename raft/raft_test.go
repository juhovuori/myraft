package raft

import (
	"testing"

	"github.com/juhovuori/myraft/comm"
)

type MockComm func(interface{}) interface{}

func (m MockComm) BroadcastRPC(request interface{}) interface{} {
	return m(request)
}

func TestRulesForAll(t *testing.T) {
	var node *SimpleRaftNode
	// • If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)
	// TODO

	// • If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower (§5.1)
	node = NewSimpleRaftNode(comm.NodeID(1), nil, RaftConfig{})
	node.state = Follower
}

func TestRulesForFollowers(t *testing.T) {
	// • Respond to RPCs from candidates and leader
	node := NewSimpleRaftNode(comm.NodeID(1), nil, RaftConfig{})
	node.state = Follower

	// • If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate: convert to candidate
	node = NewSimpleRaftNode(comm.NodeID(1), nil, RaftConfig{})
	node.state = Follower

}

func TestRulesForCandidates(t *testing.T) {
	// • On conversion to candidate, start election:
	// • Increment currentTerm
	// • Vote for self
	// • Reset election timer
	// • Send RequestVote RPCs to all other servers
	// • If votes received from majority of servers: become leader
	// • If AppendEntries RPC received from new leader: convert to follower
	// • If election timeout elapses: start new election
}

func TestRulesForLeaders(t *testing.T) {
	// • Upon election: send initial empty AppendEntries RPCs (heartbeat) to each server; repeat during idle periods to prevent election timeouts (§5.2)
	// • If command received from client: append entry to local log, respond after entry applied to state machine (§5.3)
	// • If last log index ≥ nextIndex for a follower: send AppendEntries RPC with log entries starting at nextIndex
	// • If successful: update nextIndex and matchIndex for follower (§5.3)
	// • If AppendEntries fails because of log inconsistency: decrement nextIndex and retry (§5.3)
	// • If there exists an N such that N > commitIndex, a majority of matchIndex[i] ≥ N, and log[N].term == currentTerm: set commitIndex = N (§5.3, §5.4).
}
