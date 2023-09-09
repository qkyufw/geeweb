package gee

import (
	"net/http"
	"strings"
)

// router 以map的形式存放路由映射表
type router struct {
	roots    map[string]*node       // 存储每种请求方式的Trie 树根节点
	handlers map[string]HandlerFunc // 存储每种请求方式的HandlerFunc
}

// newRouter 新建一个Router
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// parsePattern Only one * is allowed
// 拆分返回字符串组
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/") // 使用 / 对字符串进行分割形成切片

	parts := make([]string, 0) // 定义空字符串切片
	for _, item := range vs {  // 如果vs里的元素非空，就添加入parts
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' { // 如果该元素以 * 开始，则parts已经包含了所有需要匹配的路径部分，可跳出循环
				break
			}
		}
	}
	return parts
}

// addRoute 添加路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0) // 插入
	r.handlers[key] = handler
}

// getRoute method 为HTTP请求方法，path为URL路径
// 用于查找指定HTTP方法和路径对应的路由节点
func (r *router) getRoute(method string, path string) (*node, Params) {
	// 在Trie树上寻找对应的HTTP方法的根节点，如果没找到返回空
	searchParts := parsePattern(path) // 得到需要搜索的内容
	var params Params
	root, ok := r.roots[method]

	// 不存在
	if !ok {
		return nil, nil
	}

	// 继续深度优先找匹配路径的节点，找到后保存到n中
	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern) // 解析路径模式和路径参数
		for index, part := range parts { // 遍历路径模式和搜索路径切片
			if part[0] == ':' { // 遇到冒号，存入params中
				params = append(params, Param{Key: part[1:], Value: searchParts[index]})
			}
			if part[0] == '*' && len(part) > 1 { // 遇到星号，匹配路径后的元素并存入字典
				params = append(params, Param{Key: part[1:], Value: strings.Join(searchParts[index:], "/")})
				break
			}
		}
		// 存在，返回节点和参数
		return n, params
	}
	return nil, nil
}

// 用于获取指定 HTTP 请求方法下的所有叶子节点，从路由器的根节点集合中取出对应请求方法的根节点
func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

// handle 设置handle方法
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.params = &params
		key := c.Method + "-" + n.pattern
		// 将路由匹配得到的Handler添加到c.handlers列表中，执行c.Next()
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) { // 不存在路由
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
