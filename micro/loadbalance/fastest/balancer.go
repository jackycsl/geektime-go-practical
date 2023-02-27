package fastest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type Balancer struct {
	mutex    sync.RWMutex
	conns    []*conn
	lastSync time.Time
	endpoint string
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	b.mutex.RLock()
	if len(b.conns) == 0 {
		b.mutex.RUnlock()
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var res *conn
	for _, c := range b.conns {
		if res == nil || res.response > c.response {
			res = c
		}
	}
	b.mutex.RUnlock()

	return balancer.PickResult{
		SubConn: res.SubConn,
		Done: func(info balancer.DoneInfo) {
		},
	}, nil
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for con, val := range info.ReadySCs {
		conns = append(conns, &conn{
			SubConn: con,
			address: val.Address,
			// 随便设置一个默认值。当然这个默认值会对初始的负载均衡有影响
			// 不过一段时间之后就没什么影响了
			response: time.Millisecond * 100,
		})
	}
	res := &Balancer{
		conns: conns,
	}

	// 基本的思路是启动一个 goroutine 异步地去拉 prometheus 上的响应时间的数据，即调用 updateResp
	// 但是有一个很大的问题：我们这里不好怎么退出，因为没有 gRPC 不会调用 Close 方法
	// 可以考虑使用 runtime.SetFinalizer 来在 res 被回收的时候得到通知
	ch := make(chan struct{}, 1)
	runtime.SetFinalizer(res, func() {
		ch <- struct{}{}
	})
	go func() {
		ticker := time.NewTicker(b.Interval)
		for {
			select {
			case <-ticker.C:
				res.updateRespTime(b.Endpoint, b.Query)
			case <-ch:
				return
			}
		}
	}()
	return res
}

func (b *Balancer) updateRespTime(endpoint, query string) {
	httpResp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, query))
	if err != nil {
		log.Fatalln("query prometheus failed", err)
		return
	}
	decoder := json.NewDecoder(httpResp.Body)

	var resp response
	err = decoder.Decode(&resp)
	if err != nil {
		log.Fatalln("http decode failed", err)
		return
	}
	if resp.Status != "success" {
		log.Fatalln("response failed", err)
		return
	}
	for _, promRes := range resp.Data.Result {
		address, ok := promRes.Metric["address"]
		if !ok {
			return
		}

		for _, c := range b.conns {
			if c.address.Addr == address {
				ms, err := strconv.ParseInt(promRes.Value[1].(string), 10, 64)
				if err != nil {
					continue
				}
				c.response = time.Duration(ms) * time.Millisecond
			}
		}
	}
}

type Builder struct {

	// prometheus 的地址
	Endpoint string
	Query    string
	// 刷新响应时间的间隔
	Interval time.Duration
}

type conn struct {
	balancer.SubConn
	address resolver.Address
	// 响应时间
	response time.Duration
}

type response struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Result     []Result `json:"result"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}
