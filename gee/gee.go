package gee

import (
	"net/http"
)

// HandlerFunc defines the request handler used by gee
// 请求handler，给用户使用，用来定义路由映射的处理方法
type HandlerFunc func(ctx *Context)

// Engine implement the interface of ServeHTTP
// 实现 ServeHTTP ，添加路由映射表
// key由请求方法和静态路由地址构成， 如 GET-/、 POST-/hello
// 这样针对不同请求方法可以有不同的处理方法（Handler）
type Engine struct {
	router *router
}

// ServeHTTP 实现 handler 接口
// 解析请求的路径，查找路由映射表
// 如果查到就执行注册的处理方法，查不到就返回404
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}

// New is the constructor of gee.Engine
// 构造函数
func New() *Engine {
	return &Engine{router: newRouter()}
}

// 添加路由方法，方法，pattern，handler函数
// 这里的pattern可以理解为路由路径
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET 请求，会将路由和处理方法注册到映射表router中
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST 请求
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run 运行，ListenAndServe 的包装
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
