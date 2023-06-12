package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 给map[string]interface{} 取了一个别名，gee.H，构建JSON数据时会更简洁
type H map[string]interface{}

type Context struct {
	// origin objects，暂时只包括Writer和Req
	Writer http.ResponseWriter
	Req    *http.Request
	// request info，提供对Method和Path这两个常用属性的直接访问
	Path   string
	Method string
	// response info
	StatusCode int
}

// newContext 构造方法
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method, // 未设置状态码
	}
}

// PostForm 提供访问PostForm参数的方法
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
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
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
