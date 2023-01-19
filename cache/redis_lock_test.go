package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/jackycsl/geektime-go-practical/cache/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_TryLock(t *testing.T) {
	testCases := []struct {
		name string
		mock func(*gomock.Controller) redis.Cmdable

		key string

		wantErr  error
		wantLock *Lock
	}{
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key: "key1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := NewClient(tc.mock(ctrl))
			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, l.key)
			// 赋予值了
			assert.NotEmpty(t, l.value)
		})
	}
}

func TestLock_Unlock(t *testing.T) {
	// ctrl := gomock.NewController()
	// defer ctrl.Finish()
	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		key   string
		value string

		lock    *Lock
		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: context.DeadlineExceeded,
			// lock: &Lock{
			// 	key:   "key1",
			// 	value: "value1",
			// 	client: func() redis.Cmdable {
			// 		cmd := mocks.NewMockCmdable(ctrl)
			// 		return cmd
			// 	}(),
			// },
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlocked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:   "key1",
			value: "value1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &Lock{
				key:    tc.key,
				value:  tc.value,
				client: tc.mock(ctrl),
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
