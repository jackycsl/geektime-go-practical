package micro

import (
	"context"
	"time"

	"github.com/jackycsl/geektime-go-practical/micro/registry"
	"google.golang.org/grpc/resolver"
)

type grpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*grpcResolverBuilder, error) {
	return &grpcResolverBuilder{r: r, timeout: timeout}, nil
}

func (b *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}
	r.resolve()
	go r.watch()

	return r, nil
}

func (b *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	// - "dns://some_authority/foo.bar"
	//   Target{Scheme: "dns", Authority: "some_authority", Endpoint: "foo.bar"}
	// registry:///localhost:8081
	target  resolver.Target
	r       registry.Registry
	cc      resolver.ClientConn
	timeout time.Duration
	close   chan struct{}
}

func (g *grpcResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint)
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			g.resolve()
			//case event := <- events:
			//switch event.Type {
			//case "DELETE":
			//	// 删除已有的节点
			//
			//}
		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
	//g.close <- struct{}{}
}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint)
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))
	for _, si := range instances {
		address = append(address, resolver.Address{Addr: si.Address})
	}

	err = g.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}
