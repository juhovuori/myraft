package comm

import (
	"fmt"
	"time"
)

type RPCTimeout struct{}

type Comm interface {
	Joiner
	Broadcaster
	Cluster
	//RPC(NodeID, interface{}) <-chan interface{}
}

type Cluster interface {
	Count() int
}

type Joiner interface {
	Join(Node) error
}

type Broadcaster interface {
	BroadcastRPC(interface{}) <-chan interface{}
}

type MemoryComm struct {
	nodes           map[NodeID]Node
	minMessageDelay time.Duration
	maxMessageDelay time.Duration
	rpcTimeout      time.Duration
}

func NewMemoryComm() *MemoryComm {
	return &MemoryComm{
		nodes:           map[NodeID]Node{},
		minMessageDelay: time.Duration(0),
		maxMessageDelay: time.Millisecond * 100,
		rpcTimeout:      time.Millisecond * 99,
	}
}

func (c *MemoryComm) Count() int {
	return len(c.nodes)
}

func (c *MemoryComm) Join(n Node) error {
	if _, ok := c.nodes[n.ID()]; ok {
		return fmt.Errorf("Duplicate ID %v", n.ID())
	}
	c.nodes[n.ID()] = n
	return nil
}

func (c *MemoryComm) BroadcastRPC(message interface{}) <-chan interface{} {
	results := make(chan interface{})
	for _, node := range c.nodes {
		go c.rpc(node, message, results)
	}
	return results
}

/*
func (c *MemoryComm) RPC(id NodeID, message interface{}) <-chan interface{} {
	result := make(chan interface{})
	if n, ok := c.nodes[id]; ok {
		go c.rpc(n, message, result)
	} else {
		result <- fmt.Errorf("Invalid node %v", id)
	}
	return result
}
*/
func (c *MemoryComm) rpc(n Node, message interface{}, result chan<- interface{}) {
	timeout := time.After(c.rpcTimeout)
	response := make(chan interface{})
	go func() {
		time.Sleep(Delay(c.minMessageDelay, c.maxMessageDelay))
		response <- n.OnRPC(message)
	}()
	select {
	case <-timeout:
		result <- RPCTimeout{}
	case res := <-response:
		result <- res
	}
}
