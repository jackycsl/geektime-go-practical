package queue

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"unsafe"
)

var (
	ErrOutOfCapacity = errors.New("ekit: 超出最大容量限制")
	ErrEmptyQueue    = errors.New("ekit: 队列为空")
)

var _ Queue[any] = &ConcurrentLinkedQueue[any]{}

type ConcurrentLinkedQueue[T any] struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	count uint64
}

func NewConcurrentLinkedQueue[T any]() *ConcurrentLinkedQueue[T] {
	head := &node[T]{}
	ptr := unsafe.Pointer(head)
	return &ConcurrentLinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}

func (c *ConcurrentLinkedQueue[T]) EnQueue(ctx context.Context, data T) error {
	newNode := &node[T]{
		val: data,
	}
	newNodePtr := unsafe.Pointer(newNode)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		tailPtr := atomic.LoadPointer(&c.tail)

		if atomic.CompareAndSwapPointer(&c.tail, tailPtr, newNodePtr) {
			tailNode := (*node[T])(tailPtr)

			atomic.StorePointer(&tailNode.next, newNodePtr)
			atomic.AddUint64(&c.count, 1)
			return nil
		}
	}

	// 先改 tail.next
	// newNode := &node[T]{val: t}
	// newPtr := unsafe.Pointer(newNode)
	// for {
	// 	tailPtr := atomic.LoadPointer(&c.tail)
	// 	tail := (*node[T])(tailPtr)
	// 	tailNext := atomic.LoadPointer(&tail.next)
	// 	if tailNext != nil {
	// 		// 已经被人修改了，我们不需要修复，因为预期中修改的那个人会把 c.tail 指过去
	// 		continue
	// 	}
	// 	if atomic.CompareAndSwapPointer(&tail.next, tailNext, newPtr) {
	// 		// 如果失败也不用担心，说明有人抢先一步了
	// 		atomic.CompareAndSwapPointer(&c.tail, tailPtr, newPtr)
	// 		return nil
	// 	}
	// }
}

func (c *ConcurrentLinkedQueue[T]) DeQueue(ctx context.Context) (T, error) {
	for {
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		headPtr := atomic.LoadPointer(&c.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPtr)
		if head == tail {
			var t T
			return t, ErrEmptyQueue
		}
		headNextPr := atomic.LoadPointer(&head.next)
		if atomic.CompareAndSwapPointer(&c.head, headPtr, headNextPr) {
			headNext := (*node[T])(headNextPr)
			return headNext.val, nil
		}
	}
}

func (c *ConcurrentLinkedQueue[T]) IsFull() bool {
	fmt.Println()
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsEmpty() bool {
	// TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) Len() uint64 {
	return atomic.LoadUint64(&c.count)
}

type node[T any] struct {
	next unsafe.Pointer
	val  T
}
