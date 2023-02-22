package route

import (
	"context"
	"testing"
	"time"

	"github.com/jackycsl/geektime-go-practical/micro"
	"github.com/jackycsl/geektime-go-practical/micro/proto/gen"
	"github.com/jackycsl/geektime-go-practical/micro/registry/etcd"
	"github.com/jackycsl/geektime-go-practical/micro/route"
	roundrobin "github.com/jackycsl/geektime-go-practical/micro/route/round_robin"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	client, err := micro.NewClient(micro.ClientInsecure(),
		micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickerBuilder("GROUP_ROUND_ROBIN", &roundrobin.Builder{
			Filter: route.GroupFilterBuilder{}.Build()}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	require.NoError(t, err)

	ctx = context.WithValue(ctx, "group", "A")
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	uc := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 13})
		require.NoError(t, err)
		t.Log(resp)
	}

}
