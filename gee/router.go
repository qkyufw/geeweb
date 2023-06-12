package gee

import (
	"log"
	"net/http"
)

// router 以map的形式存放路由映射表
type router struct {
	handlers map[string]HandlerFunc
}

// newRouter 新建一个Router
func newRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc)}
}

// addRoute 添加路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

// handle 设置handle方法
func (r *router) handle(c *Context) {
	key := c.Method + "-" + c.Path
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
