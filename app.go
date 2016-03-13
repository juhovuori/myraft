package main

import (
	"time"

	"github.com/juhovuori/myraft/raft"
)

func main() {
	const nodeCount = 3
	comm := raft.NewMemoryComm()
	cluster := raft.NewSimpleCluster(nodeCount)
	nodes := []*raft.SimpleRaftNode{}
	for nodeID := range cluster.Iter() {
		n := raft.NewSimpleRaftNode(nodeID, comm, nodeCount)
		comm.Join(n)
		go n.Raft(comm, cluster)
		nodes = append(nodes, n)
	}
	time.Sleep(1000 * time.Millisecond)
	for _, node := range nodes {
		node.Stop()
	}
}
