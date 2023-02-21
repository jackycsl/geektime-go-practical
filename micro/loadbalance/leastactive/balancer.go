package leastactive

import (
	"math"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Balancer struct {
	connections []*activeConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	res := &activeConn{
		cnt: math.MaxInt32,
	}
	for _, c := range b.connections {
		if atomic.LoadInt32(&c.cnt) <= res.cnt {
			res = c
		}
	}
	atomic.AddInt32(&res.cnt, 1)
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(di balancer.DoneInfo) {
			atomic.AddInt32(&res.cnt, -1)
		},
	}, nil
}

type BalancerBuilder struct{}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*activeConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, &activeConn{
			c: c,
		})
	}
	return &Balancer{
		connections: connections,
	}
}

type activeConn struct {
	// 正在处理的请求数量
	cnt int32
	c   balancer.SubConn
}
