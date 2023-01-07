package queue

import (
	"log"
	"sync/atomic"
	"testing"
)

func TestCAS(t *testing.T) {
	var value int64 = 10

	res := atomic.CompareAndSwapInt64(&value, 10, 12)

	// 这个不是并发安全的，要么就是利用锁，要么就是我们刚才的 CAS
	// value = 12

	// res := atomic.CompareAndSwapInt64(&value, 11, 12)
	log.Println(res)
	log.Println(value)
}
