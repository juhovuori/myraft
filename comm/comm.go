package comm

import (
	"fmt"
	"time"
)

type RPCTimeout struct{}

type MulticastResponse struct {
	NodeID  NodeID
	Payload interface{}
}

type InvalidNode struct{}

type Comm interface {
	Joiner
	Multicaster
	Cluster
}

type Cluster interface {
	Count() int
}

type Joiner interface {
	Join(Node) error
}

type Multicaster interface {
	MulticastRPC(interface{}, ...NodeID) <-chan MulticastResponse
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
		maxMessageDelay: time.Millisecond * 50,
		rpcTimeout:      time.Millisecond * 90,
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

func (c *MemoryComm) MulticastRPC(message interface{}, destinations ...NodeID) <-chan MulticastResponse {
	results := make(chan MulticastResponse)
	for _, nodeID := range destinations {
		if node, ok := c.nodes[nodeID]; ok {
			go c.rpc(node, message, results)
		} else {
			results <- MulticastResponse{nodeID, InvalidNode{}}
		}
	}
	return results
}

func (c *MemoryComm) rpc(n Node, message interface{}, result chan<- MulticastResponse) {
	timeout := time.After(c.rpcTimeout)
	response := make(chan interface{})
	var payload interface{}
	go func() {
		time.Sleep(Delay(c.minMessageDelay, c.maxMessageDelay))
		r := n.OnRPC(message)
		time.Sleep(Delay(c.minMessageDelay, c.maxMessageDelay))
		response <- r
	}()
	select {
	case <-timeout:
		payload = RPCTimeout{}
	case res := <-response:
		payload = res
	}
	result <- MulticastResponse{n.ID(), payload}
}
