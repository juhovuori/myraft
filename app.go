package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/juhovuori/myraft/comm"
	"github.com/juhovuori/myraft/raft"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Usage myraft <number-of-nodes>")
		os.Exit(1)
	}
	nodeCount, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	runRaft(nodeCount)
}

func runRaft(nodeCount int) {
	c := comm.NewMemoryComm()

	// Create config
	config := raft.RaftConfig{}
	for i := 1; i <= nodeCount; i++ {
		nodeID := comm.NodeID(fmt.Sprintf("1.2.3.%d", i))
		config.Nodes = append(config.Nodes, nodeID)
	}

	// Create nodes
	nodes := []*raft.SimpleRaftNode{}
	for _, nodeID := range config.Nodes {
		n := raft.NewSimpleRaftNode(nodeID, c, config)
		c.Join(n)
		nodes = append(nodes, n)
	}

	// Start nodes
	for _, node := range nodes {
		go node.Start()
	}

	// Send random queries for some time
	presumedLeader := config.Nodes[0]
	t1 := time.After(3 * time.Second)
	for timeout := false; !timeout; {
		t2 := time.After(comm.Delay(0, 100*time.Millisecond))
		select {
		case <-t1:
			timeout = true
			break
		case <-t2:
			msg := raft.Command(string(presumedLeader))
			response := <-c.MulticastRPC(msg, presumedLeader)
			cmdResponse, ok := response.Payload.(raft.CommandResult)
			if ok {
				presumedLeader = cmdResponse.Leader
			}
		}
	}

	// Stop nodes
	for _, node := range nodes {
		node.Stop()
		node.DumpLog()
	}
}
