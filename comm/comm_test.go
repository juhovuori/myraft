package comm

import "testing"

func TestJoin(t *testing.T) {
	comm := NewMemoryComm()
	id := NodeID(1)
	node := NewSimpleNode(id)
	comm.Join(node)
	n, ok := comm.nodes[id]
	if !ok {
		t.Error("Node not found")
	} else if n != node {
		t.Error("Wrong node")
	}
}
