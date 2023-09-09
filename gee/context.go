package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 给map[string]interface{} 取了一个别名，gee.H，构建JSON数据时会更简洁
type H map[string]interface{}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = -1
	w.status = http.StatusOK
}

type Context struct {
	// origin objects，暂时只包括Writer和Req
	writermem responseWriter // 增加的属性，用来提供对Writer的访问
	Writer    http.ResponseWriter
	Req       *http.Request
	// request info，提供对Method和Path这两个常用属性的直接访问
	Path   string
	Method string
	params *Params // 增加的属性，用来提供对路由参数的访问
	// response info
	StatusCode int
	// middleware
	handlers HandlersChain // 中间件列表
	index    int
	engine   *Engine
}

// newContext 构造方法
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method, // 未设置状态码
		index:  -1,         // 调用顺序，记录执行到第几个中间件
	}
}

// Next 用于调用下一个中间件，调用完成后，执行本中间件Next()函数之后未执行的部分
func (c *Context) Next() {
	c.index++            // index 后移
	s := len(c.handlers) // 长度
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c) // 按index顺序调用方法
	}
}

func (c *Context) Param(key string) string {
	for _, entry := range *c.params {
		if entry.Key == key {
			return entry.Value
		}
	}
	return ""
}

// PostForm 提供访问PostForm参数的方法
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// Query 提供访问Query参数的方法
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Status 设置状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 的响应方法
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 的响应方法
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		// err != nil 时，下面的内容不执行，因为前面已经执行了WriteHeader(code)
		// 返回码将不会更改，旗下内容操作将无效
		// encoder.Encode相当于调用了Write()
		// http.Error里的WriteHeader、Set操作均无效
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 的响应方法
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write([]byte(data))
}

// HTML 的响应方法
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) reset() {
	c.Writer = &c.writermem
	c.handlers = nil
	c.index = -1

	*c.params = (*c.params)[:0]
}
