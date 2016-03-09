package raft

import (
	"fmt"
	"log"
)

type State string
type Term int

const (
	Follower  = State("follower")
	Candidate = State("candidate")
	Leader    = State("leader")
)

type RaftNode interface {
	Node
	StartRaft(comm Comm, cluster Cluster)
	StopRaft()
}

type SimpleRaftNode struct {
	nodeID       NodeID
	nodeIDPretty string
	state        State
	term         Term
	hbTimer      HeartbeatTimer
}

func NewSimpleRaftNode(nodeID NodeID) *SimpleRaftNode {
	nodeIDPretty := fmt.Sprintf("%03d", nodeID)
	node := SimpleRaftNode{
		nodeID,
		nodeIDPretty,
		Follower,
		Term(0),
		nil,
	}
	node.hbTimer = NewDefaultTimer(&node)
	return &node
}
func (n *SimpleRaftNode) StartRaft(comm Comm, cluster Cluster) {
	n.hbTimer.Reset()
	n.log("Started raft")
}

func (n *SimpleRaftNode) StopRaft() {
	n.log("Stopped raft")
}

func (n *SimpleRaftNode) OnHeartbeatTimeout() {
	n.log("Heartbeat timeout")
}

func (n *SimpleRaftNode) OnMessage(m Message) {
	n.log("Got message", m)
}

func (n *SimpleRaftNode) ID() NodeID {
	return n.nodeID
}

func (n *SimpleRaftNode) log(msgs ...interface{}) {
	msgs = append([]interface{}{n.nodeIDPretty}, msgs...)
	log.Println(msgs...)
}
