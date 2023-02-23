package roundrobin

import (
	"math"
	"sync"

	"github.com/jackycsl/geektime-go-practical/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type WeightBalancer struct {
	connections []*weightConn
	filter      route.Filter
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var totalWeight uint32
	var res *weightConn

	for _, c := range w.connections {
		if w.filter != nil && !w.filter(info, c.addr) {
			continue
		}
		c.mutex.Lock()
		totalWeight = totalWeight + c.efficientWeight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if res == nil || res.currentWeight < c.currentWeight {
			res = c
		}
		c.mutex.Unlock()
	}

	if res == nil {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res.mutex.Lock()
	res.currentWeight = res.currentWeight - totalWeight
	res.mutex.Unlock()

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			res.mutex.Lock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil && res.efficientWeight == math.MaxUint32 {
				return
			}
			if info.Err != nil {
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
			res.mutex.Unlock()
		},
	}, nil
}

type WeightBalancerBuilder struct {
	Filter route.Filter
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		weight := subInfo.Address.Attributes.Value("weight").(uint32)

		cs = append(cs, &weightConn{
			c:               sub,
			weight:          uint32(weight),
			currentWeight:   uint32(weight),
			efficientWeight: uint32(weight),
			addr:            subInfo.Address,
		})
	}
	return &WeightBalancer{
		connections: cs,
		filter:      w.Filter,
	}
}

type weightConn struct {
	mutex           sync.Mutex
	c               balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
	addr            resolver.Address
}
