package random

import (
	"math/rand"

	"github.com/jackycsl/geektime-go-practical/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type WeightBalancer struct {
	connections []*weightConn
	filter      route.Filter
}

func (b *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var totalWeight uint32
	candidates := make([]*weightConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
		totalWeight = totalWeight + c.weight
	}

	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	tgt := rand.Intn(int(totalWeight) + 1)
	var idx int
	for i, c := range candidates {
		tgt = tgt - int(c.weight)
		if tgt <= 0 {
			idx = i
			break
		}
	}
	return balancer.PickResult{
		SubConn: candidates[idx].c,
		Done:    func(di balancer.DoneInfo) {},
	}, nil
}

type WeightBalancerBuilder struct {
	Filter route.Filter
}

func (b *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		cs = append(cs, &weightConn{
			c:      sub,
			weight: weight,
			addr:   subInfo.Address,
		})
	}
	return &WeightBalancer{
		connections: cs,
		filter:      b.Filter,
	}
}

type weightConn struct {
	c      balancer.SubConn
	weight uint32
	addr   resolver.Address
}
