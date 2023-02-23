package random

import (
	"math/rand"

	"github.com/jackycsl/geektime-go-practical/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type Balancer struct {
	connections []subConn
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	idx := rand.Intn(len(candidates))
	return balancer.PickResult{
		SubConn: candidates[idx].c,
		Done:    func(di balancer.DoneInfo) {},
	}, nil
}

type BalancerBuilder struct {
	Filter route.Filter
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			addr: ci.Address,
		})
	}
	return &Balancer{
		connections: connections,
		filter:      b.Filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	addr resolver.Address
}
