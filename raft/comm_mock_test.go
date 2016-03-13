package raft

import (
	"fmt"
)

type MockComm struct {
	nodes    map[NodeID]Node
	messages map[NodeID][]Message
}

func NewMockComm() *MockComm {
	return &MockComm{map[NodeID]Node{}, map[NodeID][]Message{}}
}

func (c *MockComm) Join(n Node) error {
	if _, ok := c.nodes[n.ID()]; ok {
		return fmt.Errorf("Duplicate ID %v", n.ID())
	}
	c.nodes[n.ID()] = n
	return nil
}

func (c *MockComm) Broadcast(message Message) error {
	for id, _ := range c.nodes {
		if err := c.Send(id, message); err != nil {
			return err
		}
	}
	return nil
}

func (c *MockComm) Send(id NodeID, message Message) error {
	c.messages[id] = append(c.messages[id], message)
	if _, ok := c.nodes[id]; !ok {
		return fmt.Errorf("Invalid node %v", id)
	}
	return nil
}
