package raft

import (
	"fmt"
)

type Message string

type Comm interface {
	Join(Node) error
	Send(NodeID, Message)
}

type MemoryComm struct {
	nodes map[NodeID]Node
}

func NewMemoryComm() *MemoryComm {
	return &MemoryComm{}
}
func (c *MemoryComm) Join(n Node) error {
	_, ok := c.nodes[n.ID()]
	if ok {
		return fmt.Errorf("Duplicate ID %v", n.ID())
	}
	c.nodes[n.ID()] = n
	return nil
}

func (c *MemoryComm) Send(NodeID, Message) {
}
