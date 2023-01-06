package queue

import "context"

type Queue[T any] interface {
	EnQueue(ctx context.Context, data T) error
	DeQueue(ctx context.Context) (T, error)

	IsFull() bool
	IsEmpty() bool
	Len() uint64
}
