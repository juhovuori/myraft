package raft

import (
	"testing"
)

func createRaftNode(state State) (*SimpleRaftNode, *MockComm) {
	comm := NewMockComm()
	node := NewSimpleRaftNode(NodeID(1), comm, 1)
	node.state = state
	comm.Join(node)
	return node, comm
}

func TestRulesForAll(t *testing.T) {
	var node *SimpleRaftNode
	// • If commitIndex > lastApplied: increment lastApplied, apply log[lastApplied] to state machine (§5.3)
	// TODO

	// • If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower (§5.1)
	node, _ = createRaftNode(Follower)
	node.RunCommand(RequestVote{4, NodeID(2), 0, 3})
	if node.state != Follower || node.currentTerm != 4 {
		t.Error(node.state, node.currentTerm)
	}
}

func TestRulesForFollowers(t *testing.T) {
	// • Respond to RPCs from candidates and leaders
	node, comm := createRaftNode(Follower)
	sender := NodeID(2)
	node.RunCommand(RequestVote{4, sender, 0, 3})
	if len(comm.messages[sender]) != 1 {
		t.Error(comm.messages)
	}

	// • If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate: convert to candidate
	node, comm = createRaftNode(Follower)
	node.RunCommand(ElectionTimeout{})
	if node.state != Candidate {
		t.Error(node.state)
	}

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
