package util

import "sync"

type Counter struct {
	counter int

	lock sync.Mutex
}

func NewCounter() *Counter {
	return &Counter{1, sync.Mutex{}}
}

func (c *Counter) Next() int {
	c.lock.Lock()
	res := c.counter
	defer c.lock.Unlock()
	c.counter++
	return res
}
