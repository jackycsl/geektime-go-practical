package accesslog

import (
	"encoding/json"

	"github.com/jackycsl/geektime-go-practical/web/v7"
)

type MiddlewareBuilder struct {
	logFunc func(log string)
}

func (m *MiddlewareBuilder) LogFunc(fn func(log string)) *MiddlewareBuilder {
	m.logFunc = fn
	return m
}

func (m *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				l := accesslog{
					Host:       ctx.Req.Host,
					Route:      ctx.MatchedRoute,
					HTTPMethod: ctx.Req.Method,
					Path:       ctx.Req.URL.Path,
				}
				data, _ := json.Marshal(l)
				m.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

type accesslog struct {
	Host       string `json:"host,omitempty"`
	Route      string `json:"route,omitempty"`
	HTTPMethod string `json:"http_method,omitempty"`
	Path       string `json:"path,omitempty"`
}
