package raft

type NodeID int

type Node interface {
	OnMessage(Message)
	ID() NodeID
}
