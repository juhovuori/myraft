package main

import (
	"fmt"
	"time"

	"github.com/juhovuori/myraft/comm"
	"github.com/juhovuori/myraft/raft"
)

func main() {
	const nodeCount = 3
	c := comm.NewMemoryComm()

	// Create nodes
	nodes := []*raft.SimpleRaftNode{}
	for i := 1; i < 4; i++ {
		nodeID := comm.NodeID(fmt.Sprintf("1.2.3.%d", i))
		n := raft.NewSimpleRaftNode(nodeID, c, nodeCount)
		c.Join(n)
		nodes = append(nodes, n)
	}

	// Start nodes
	for _, node := range nodes {
		go node.Start()
	}

	//
	time.Sleep(1000 * time.Millisecond)

	// Stop nodes
	for _, node := range nodes {
		node.Stop()
	}
}
