package cache

import (
	"context"
	"testing"
	"time"

	"github.com/gotomicro/ekit/list"
	"github.com/stretchr/testify/assert"
)

func TestMaxMemoryCache_Get(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string

		wantKeys []string
		wantErr  error
	}{
		{
			name: "not exist",
			cache: func() *MaxMemoryCache {
				return NewMaxMemoryCache(100, &mockCache{})
			},
			key:      "key1",
			wantKeys: []string{},
			wantErr:  errKeyNotFound,
		},
		{
			name: "exist",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"key1": []byte("OK"),
						"key2": []byte("OK"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1", "key2"})
				return res
			},
			key:      "key1",
			wantKeys: []string{"key2", "key1"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			_, err := cache.Get(context.Background(), tc.key)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantKeys, cache.keys.AsSlice())
		})
	}
}

func TestMaxMemoryCache_Set(t *testing.T) {
	testCases := []struct {
		name  string
		cache func() *MaxMemoryCache

		key string
		val []byte

		wantKeys []string
		wantErr  error
		wantUsed int64
	}{
		{
			name: "not exist",
			cache: func() *MaxMemoryCache {
				return NewMaxMemoryCache(100, &mockCache{data: map[string][]byte{}})
			},
			key:      "key1",
			val:      []byte("hello"),
			wantKeys: []string{"key1"},
			wantUsed: 5,
		},
		{
			name: "add new",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1"})
				res.used = 5
				return res
			},
			key:      "key2",
			val:      []byte("world"),
			wantKeys: []string{"key1", "key2"},
			wantUsed: 10,
		},
		{
			name: "override-incr",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1"})
				res.used = 5
				return res
			},
			key:      "key1",
			val:      []byte("hello,world"),
			wantKeys: []string{"key1"},
			wantUsed: 11,
		},
		{
			name: "override-decr",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(100, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1"})
				res.used = 5
				return res
			},
			key:      "key1",
			val:      []byte("he"),
			wantKeys: []string{"key1"},
			wantUsed: 2,
		},
		{
			name: "delete",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello, key1"),
						"key2": []byte("hello, key2"),
						"key3": []byte("hello, key3"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1", "key2", "key3"})
				res.used = 33
				return res
			},
			key:      "key4",
			val:      []byte("hello, key4"),
			wantKeys: []string{"key2", "key3", "key4"},
			wantUsed: 33,
		},
		{
			name: "delete-multi",
			cache: func() *MaxMemoryCache {
				res := NewMaxMemoryCache(40, &mockCache{
					data: map[string][]byte{
						"key1": []byte("hello, key1"),
						"key2": []byte("hello, key2"),
						"key3": []byte("hello, key3"),
					},
				})
				res.keys = list.NewLinkedListOf([]string{"key1", "key2", "key3"})
				res.used = 33
				return res
			},
			key:      "key4",
			val:      []byte("hello, key4,hello, key4"),
			wantKeys: []string{"key3", "key4"},
			wantUsed: 34,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache()
			err := cache.Set(context.Background(), tc.key, tc.val, time.Minute)
			assert.Equal(t, tc.wantKeys, cache.keys.AsSlice())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUsed, cache.used)
		})
	}
}

type mockCache struct {
	Cache
	fn   func(key string, val []byte)
	data map[string][]byte
}

func (m *mockCache) Set(ctx context.Context, key string, val []byte, expiration time.Duration) error {
	m.data[key] = val
	return nil
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		return val, nil
	}
	return nil, errKeyNotFound
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	val, ok := m.data[key]
	if ok {
		m.fn(key, val)
	}
	return nil
}

func (m *mockCache) LoadAndDelete(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.data[key]
	if ok {
		m.fn(key, val)
		return val, nil
	}
	return nil, errKeyNotFound
}

func (m *mockCache) OnEvicted(fn func(key string, val []byte)) {
	m.fn = fn
}
