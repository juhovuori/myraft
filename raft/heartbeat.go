package raft

import (
	"time"
)

type HeartbeatTimer interface {
	Reset()
}

type DefaultTimer struct {
	hbTimeout time.Duration
	round     int
	timeout   chan int
	delegate  TimeoutListener
}

func NewDefaultTimer(delegate TimeoutListener) *DefaultTimer {
	t := DefaultTimer{
		300 * time.Millisecond,
		0,
		make(chan int),
		delegate,
	}
	t.Reset()
	go t.wakeup()
	return &t
}

func (t *DefaultTimer) Reset() {
	t.round++
	go t.alarm(t.round)
}

func (t *DefaultTimer) alarm(round int) {
	time.Sleep(t.hbTimeout)
	t.timeout <- round
}

func (t *DefaultTimer) wakeup() {
	for round := range t.timeout {
		if round == t.round {
			t.delegate.OnHeartbeatTimeout()
		}
	}
}

type TimeoutListener interface {
	OnHeartbeatTimeout()
}
