package web

import (
	"fmt"
	"net/http"
)

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	// Start 启动服务器
	// addr 是监听地址。如果只指定端口，可以使用 ":8081"
	// 或者 "localhost:8082"
	Start(addr string) error

	// addRoute 注册一个路由
	// method 是 HTTP 方法
	addRoute(method string, path string, handler HandleFunc)
	// 我们并不采取这种设计方案
	// addRoute(method string, path string, handlers... HandleFunc)
}

// 确保 HTTPServer 肯定实现了 Server 接口
var _ Server = &HTTPServer{}

type HTTPServerOption func(server *HTTPServer)

type HTTPServer struct {
	router
	mdls      []Middleware
	log       func(msg string, args ...any)
	tplEngine TemplateEngine
}

func NewHTTPServer(opts ...HTTPServerOption) *HTTPServer {
	res := &HTTPServer{
		router: newRouter(),
		log: func(msg string, args ...any) {
			fmt.Printf(msg, args...)
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ServerWithTemplateEngine(tplEngine TemplateEngine) HTTPServerOption {
	return func(server *HTTPServer) {
		server.tplEngine = tplEngine
	}
}

func ServerWithMiddleware(mdls ...Middleware) HTTPServerOption {
	return func(server *HTTPServer) {
		server.mdls = mdls
	}
}

// ServeHTTP HTTPServer 处理请求的入口
func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:       request,
		Resp:      writer,
		tplEngine: s.tplEngine,
	}
	// 最后一个应该是 HTTPServer 执行路由匹配，执行用户代码
	root := s.serve
	// 从后往前组装
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}

	// 这里，最后一步，把RespData 和 RespStatusCode 刷新到响应里面
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			s.flashResp(ctx)
		}
	}
	root = m(root)
	root(ctx)
}

func (s *HTTPServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode != 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	n, err := ctx.Resp.Write(ctx.RespData)
	if err != nil || n != len(ctx.RespData) {
		s.log("写入响应失败 %v", err)
	}
}

// Start 启动服务器
func (s *HTTPServer) Start(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *HTTPServer) Post(path string, handler HandleFunc) {
	s.addRoute(http.MethodPost, path, handler)
}

func (s *HTTPServer) Get(path string, handler HandleFunc) {
	s.addRoute(http.MethodGet, path, handler)
}

func (s *HTTPServer) serve(ctx *Context) {
	mi, ok := s.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || mi.n == nil || mi.n.handler == nil {
		ctx.RespStatusCode = 404
		ctx.RespData = []byte("NOT FOUND")
		return
	}
	ctx.PathParams = mi.pathParams
	ctx.MatchedRoute = mi.n.route
	mi.n.handler(ctx)
}
