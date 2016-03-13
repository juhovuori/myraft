package raft

import (
	"fmt"
	"time"
)

type Comm interface {
	Join(Node) error
	Send(NodeID, Message) error
	Broadcast(Message) error
}

type MemoryComm struct {
	nodes map[NodeID]Node
}

func NewMemoryComm() *MemoryComm {
	return &MemoryComm{map[NodeID]Node{}}
}
func (c *MemoryComm) Join(n Node) error {

	if _, ok := c.nodes[n.ID()]; ok {
		return fmt.Errorf("Duplicate ID %v", n.ID())
	}
	c.nodes[n.ID()] = n
	return nil
}

func (c *MemoryComm) Broadcast(message Message) error {
	for id, _ := range c.nodes {
		if err := c.Send(id, message); err != nil {
			return err
		}
	}
	return nil
}

func (c *MemoryComm) Send(id NodeID, message Message) error {
	n, ok := c.nodes[id]
	if !ok {
		return fmt.Errorf("Invalid node %v", id)
	}
	go c.send(n, message)
	return nil
}

func (c *MemoryComm) send(n Node, message Message) {
	time.Sleep(Delay(0, time.Millisecond*100))
	n.OnMessage(message)
}
