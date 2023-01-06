package queue

import (
	"context"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex   *sync.Mutex
	data    []T
	maxSize int

	notFull  chan struct{}
	notEmpty chan struct{}
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:     make([]T, maxSize),
		mutex:    m,
		notFull:  make(chan struct{}, 1),
		notEmpty: make(chan struct{}, 1),
		maxSize:  maxSize,
	}
}

func (c *ConcurrentBlockingQueue[T]) EnQueue(ctx context.Context, data T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	c.mutex.Lock()

	for c.IsFull() {
		c.mutex.Unlock()
		select {
		case <-c.notFull:
			c.mutex.Lock()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.data = append(c.data, data)
	if len(c.data) == 1 {
		c.notEmpty <- struct{}{}
	}
	c.mutex.Unlock()

	return nil
}

func (c *ConcurrentBlockingQueue[T]) DeQueue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}

	c.mutex.Lock()

	for c.IsEmpty() {
		c.mutex.Unlock()
		select {
		case <-c.notEmpty:
			c.mutex.Lock()
		case <-ctx.Done():
			var t T
			return t, ctx.Err()
		}
	}
	t := c.data[0]
	c.data = c.data[1:]
	if len(c.data) == c.maxSize-1 {
		c.notFull <- struct{}{}
	}
	c.mutex.Unlock()

	return t, nil
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentBlockingQueue[T]) isFull() bool {
	return len(c.data) == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	return uint64(len(c.data))
}
