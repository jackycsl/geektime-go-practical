package web

import (
	"fmt"
	"testing"
)

// 这里放着端到端测试的代码

func TestServer(t *testing.T) {
	s := NewHTTPServer()
	s.Get("/", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, world"))
	})
	s.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello, user"))
	})
	s.Post("/form", func(ctx *Context) {
		ctx.Req.ParseForm()
		ctx.Resp.Write([]byte(fmt.Sprintf("hello %s", ctx.Req.URL.Path)))
	})
	s.Get("/values/:id", func(ctx *Context) {
		id, err := ctx.PathValue("id").ToInt64()
		if err != nil {
			ctx.Resp.WriteHeader(400)
			ctx.Resp.Write([]byte("id 输入不对"))
		}
		ctx.Resp.Write([]byte(fmt.Sprintf("hello %d", id)))
	})

	type User struct {
		Name string `json:"name"`
	}

	s.Get("/user/123", func(ctx *Context) {
		ctx.RespJSON(202, User{
			Name: "Tom",
		})
	})

	s.Start(":8081")
}
