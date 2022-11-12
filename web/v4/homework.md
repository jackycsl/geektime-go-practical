# 选做

基准测试结果

- 静态匹配
- 通配符匹配
- 路径参数
- 正则匹配

```bash
✗ go test -bench=. -run=^$
goos: darwin
goarch: arm64
pkg: github.com/jackycsl/geektime-go-practical/web/v4
BenchmarkFindRoute_Static-8             17445517                69.38 ns/op
BenchmarkFindRoute_Wildcard-8           16770882                70.32 ns/op
BenchmarkFindRoute_Param-8               6978526               172.5 ns/op
BenchmarkFindRoute_Regex-8               4740439               267.0 ns/op
PASS
ok      github.com/jackycsl/geektime-go-practical/web/v4        6.968s
```

从结果看来，路径参数和正则匹配花的时间较久。
我们还可以利用 benchmark 生成的 profile 文件，分析路由树的瓶颈。

```bash
go test -bench=BenchmarkFindRoute_Regex -run=^$  -cpuprofile regex_profile.out
```

接下来

```bash
✗ go tool pprof regex_profile.out
Type: cpu
Time: Nov 12, 2022 at 9:21am (+08)
Duration: 1.61s, Total samples = 1.57s (97.64%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) list findRoute
Total: 1.57s
ROUTINE ========================
      10ms      170ms (flat, cum) 10.83% of Total
         .          .     66:   root.handler = handler
         .          .     67:}
         .          .     68:
         .          .     69:// findRoute 查找对应的节点
         .          .     70:// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
      10ms       10ms     71:func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
         .          .     72:   root, ok := r.trees[method]
         .          .     73:   if !ok {
         .          .     74:           return nil, false
         .          .     75:   }
         .          .     76:
         .          .     77:   if path == "/" {
         .          .     78:           return &matchInfo{n: root}, true
         .          .     79:   }
         .       30ms     80:   segs := strings.Split(strings.Trim(path, "/"), "/")
         .       10ms     81:   mi := &matchInfo{}
         .          .     82:   for _, s := range segs {
         .          .     83:           var matchParam bool
         .       30ms     84:           root, matchParam, ok = root.childOf(s)
         .          .     85:           if !ok {
         .          .     86:                   return nil, false
         .          .     87:           }
         .          .     88:           if matchParam {
         .          .     89:                   // 正则匹配
         .          .     90:                   if strings.Contains(root.path, "(") {
         .       90ms     91:                           mi.addValue(root.path[1:strings.Index(root.path, "(")], s)
         .          .     92:                   } else {
         .          .     93:                           mi.addValue(root.path[1:], s)
         .          .     94:                   }
         .          .     95:           }
         .          .     96:   }
(pprof)
```

我们还可以生成graph来分析。

```bash
(pprof) web
```

![alt text](pprof-BenchmarkFindRoute-Regex.png 'pprof')

strings 的操作占用了较多的时间，比如split和index。如果要优化，应该可以从这下手。
