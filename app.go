package main

import (
	"time"

	"github.com/juhovuori/myraft/raft"
)

func main() {
	comm := raft.NewMemoryComm()
	cluster := raft.NewSimpleCluster(7)
	nodes := []*raft.SimpleRaftNode{}
	for nodeID := range cluster.Iter() {
		n := raft.NewSimpleRaftNode(nodeID)
		n.StartRaft(comm, cluster)
		nodes = append(nodes, n)
	}
	time.Sleep(1000 * time.Millisecond)
	for _, node := range nodes {
		node.StopRaft()
	}
}
