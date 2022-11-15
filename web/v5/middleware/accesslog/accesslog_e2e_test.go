package accesslog

import (
	"fmt"
	"testing"

	"github.com/jackycsl/geektime-go-practical/web/v5"
)

func TestMiddlewareBuilderE2E(t *testing.T) {
	builder := MiddlewareBuilder{}
	mdl := builder.LogFunc(func(log string) {
		fmt.Println(log)
	}).Build()
	server := web.NewHTTPServer(web.ServerWithMiddleware(mdl))
	server.Get("/a/b/*", func(ctx *web.Context) {
		fmt.Println("hello, it's me")
		ctx.Resp.Write([]byte("hello, it's me"))
	})
	server.Start(":8081")
}
