package comm

import (
	"fmt"
	"sync/atomic"
)

type NodeID string

type RPCReceived struct{}

type Node interface {
	OnRPC(interface{}) interface{}
	ID() NodeID
}

type SimpleNode struct {
	nodeID NodeID
	prefix string
	suffix string
}

var nodeN int32

func NewSimpleNode(nodeID NodeID) *SimpleNode {
	atomic.AddInt32(&nodeN, 1)
	prefix := fmt.Sprintf("\x1B[1;3%dm", int(nodeN)%8)
	suffix := "\x1B[0m"
	return &SimpleNode{
		nodeID,
		prefix,
		suffix,
	}
}

func (n *SimpleNode) OnRPC(msg interface{}) interface{} {
	return RPCReceived{}
}

func (n *SimpleNode) ID() NodeID {
	return n.nodeID
}

func (n *SimpleNode) Log(msgs ...interface{}) {
	msgs = append(append([]interface{}{n.prefix}, msgs...), n.suffix)
	fmt.Println(msgs...)
}

func (n *SimpleNode) Logf(f string, msgs ...interface{}) {
	fmt.Println(n.prefix, fmt.Sprintf(f, msgs...), n.suffix)
}
