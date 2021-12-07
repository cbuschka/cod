package engine

import "sync/atomic"

type Counter struct {
	curr uint64
}

func NewCounter() *Counter {
	return &Counter{0}
}

func (counter *Counter) Next() uint64 {
	return atomic.AddUint64(&counter.curr, 1)
}
