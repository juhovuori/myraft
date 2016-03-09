package raft

type Cluster interface {
	Iter() <-chan NodeID
}

type SimpleCluster struct {
	NumberOfNodes int
}

func NewSimpleCluster(nodes int) *SimpleCluster {
	return &SimpleCluster{nodes}
}

func (s *SimpleCluster) Iter() <-chan NodeID {
	c := make(chan NodeID)
	go func() {
		for i := 0; i < s.NumberOfNodes; i++ {
			c <- NodeID(i)
		}
		close(c)
	}()
	return c
}
