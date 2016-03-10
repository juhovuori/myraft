package raft

import (
	"fmt"
	"log"
)

type NodeID int

type Node interface {
	OnMessage(Message)
	ID() NodeID
}

type SimpleNode struct {
	nodeID       NodeID
	nodeIDPretty string
}

func NewSimpleNode(nodeID NodeID) *SimpleNode {
	return &SimpleNode{
		nodeID,
		fmt.Sprintf("%03d", nodeID),
	}
}

func (n *SimpleRaftNode) ID() NodeID {
	return n.nodeID
}

func (n *SimpleRaftNode) Log(msgs ...interface{}) {
	msgs = append([]interface{}{n.nodeIDPretty}, msgs...)
	log.Println(msgs...)
}

func (n *SimpleRaftNode) Logf(f string, msgs ...interface{}) {
	log.Println(n.nodeIDPretty, fmt.Sprintf(f, msgs...))
}
