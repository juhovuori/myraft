package comm

import (
	"fmt"
	"sync/atomic"
)

type NodeID string

type Node interface {
	OnRPC(interface{}) interface{}
	ID() NodeID
}

type SimpleNode struct {
	nodeID     NodeID
	nodePrefix string
}

var nodeN int32

func NewSimpleNode(nodeID NodeID) *SimpleNode {
	prefix := ""
	for i := 0; i < int(nodeN)%7; i++ {
		prefix = prefix + "                                            "
	}
	atomic.AddInt32(&nodeN, 1)
	return &SimpleNode{
		nodeID,
		prefix,
	}
}

func (n *SimpleNode) OnRPC(msg interface{}) interface{} {
	return fmt.Errorf("Not implemented")
}

func (n *SimpleNode) ID() NodeID {
	return n.nodeID
}

func (n *SimpleNode) Log(msgs ...interface{}) {
	msgs = append([]interface{}{n.nodePrefix}, msgs...)
	fmt.Println(msgs...)
}

func (n *SimpleNode) Logf(f string, msgs ...interface{}) {
	fmt.Println(n.nodePrefix, fmt.Sprintf(f, msgs...))
}
