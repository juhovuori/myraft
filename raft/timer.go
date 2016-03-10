package raft

import (
	"time"
)

var (
	ElectionMinTimeout   = 300 * time.Millisecond
	ElectionMaxTimeout   = 500 * time.Millisecond
	LeadershipMinTimeout = 150 * time.Millisecond
	LeadershipMaxTimeout = 150 * time.Millisecond
)

type Timer interface {
	Reset()
}

type DefaultTimer struct {
	minTimeout time.Duration
	maxTimeout time.Duration
	round      int
	timeout    chan int
	delegate   Delegate
	logger     Logger
}

func NewDefaultTimer(minTimeout, maxTimeout time.Duration, delegate Delegate, logger Logger) *DefaultTimer {
	t := DefaultTimer{
		minTimeout,
		maxTimeout,
		0,
		make(chan int),
		delegate,
		logger,
	}
	go t.wakeup()
	return &t
}

func (t *DefaultTimer) Reset() {
	t.round++
	go t.alarm(t.round)
}

func (t *DefaultTimer) alarm(round int) {
	d := Delay(t.minTimeout, t.maxTimeout)
	t.logger.Log("Timing out for", d)
	time.Sleep(d)
	t.timeout <- round
}

func (t *DefaultTimer) wakeup() {
	for round := range t.timeout {
		if round == t.round {
			t.delegate.OnTimeout()
		}
	}
}

type Delegate interface {
	OnTimeout()
}

type Logger interface {
	Log(msgs ...interface{})
}

type TimerCallback func()

func (cb TimerCallback) OnTimeout() {
	cb()
}
