package comm

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	comm := NewMemoryComm()

	if comm.Count() != 0 {
		t.Error("Wrong node count")
	}

	id1 := NodeID("1")
	n1 := NewSimpleNode(id1)
	join(t, comm, n1, 1)

	id2 := NodeID("2")
	n2 := NewSimpleNode(id2)
	join(t, comm, n2, 2)

	responses := comm.BroadcastRPC("hello")
	timeout := time.After(1 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case <-timeout:
			t.Fatal("timeout")
		case <-responses:
			t.Log("got response")
		}
	}
}

func join(t *testing.T, comm *MemoryComm, node Node, nodeCount int) {
	comm.Join(node)
	n, ok := comm.nodes[node.ID()]
	if !ok {
		t.Error("Node not found")
	} else if n != node {
		t.Error("Wrong node")
	}
	if comm.Count() != nodeCount {
		t.Error("Wrong node count")
	}
}
