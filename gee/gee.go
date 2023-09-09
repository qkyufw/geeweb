package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc defines the request handler used by gee
// 请求handler，给用户使用，用来定义路由映射的处理方法
type HandlerFunc func(ctx *Context)
type HandlersChain []HandlerFunc

type (
	RouterGroup struct {
		prefix      string        // 前缀
		middlewares HandlersChain // support middleware
		parent      *RouterGroup  // support nesting，知道父分组，进行分组嵌套
		engine      *Engine       // all groups share an Engine instance，需要有访问Router的能力，所以也指向一个Engine
	}
	// Engine implement the interface of ServeHTTP
	// 实现 ServeHTTP ，添加路由映射表
	// key由请求方法和静态路由地址构成， 如 GET-/、 POST-/hello
	// 这样针对不同请求方法可以有不同的处理方法（Handler）
	// 协调整个框架资源，也可以协调不同分组
	Engine struct {
		*RouterGroup
		router        *router
		groups        []*RouterGroup     // store all groups
		htmlTemplates *template.Template // for html render，加载所有的模板
		funcMap       template.FuncMap   // for html render，所有自定义函数和加载模板的方法
	}
)

// ServeHTTP 实现 handler 接口
// 解析请求的路径，查找路由映射表
// 如果查到就执行注册的处理方法，查不到就返回404
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares HandlersChain
	for _, group := range engine.groups { // 判断请求适用于那些中间件，这里通过URL前缀盘点
		if strings.HasPrefix(req.URL.Path, group.prefix) { // 得到中间件列表后复制给c.handler
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = middlewares // 将中间件假如到handler中
	c.engine = engine
	engine.router.handle(c) // 找到该路由的Handler，添加入handlers
}

// Default 默认使用其中的两个中间件
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// Use 给group添加中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// New is the constructor of gee.Engine
// 构造函数
func New() *Engine {
	engine := &Engine{router: newRouter()}             // 初始化一个分组对象
	engine.RouterGroup = &RouterGroup{engine: engine}  // 给组内绑定engine为本engine
	engine.groups = []*RouterGroup{engine.RouterGroup} // 将该分组加入切片中
	return engine                                      // 返回engine
}

// Group 用于创建一个新的RouterGroup
// 所有的分组都是用同一个存在的engine，除了中间件外，基本都进行了设置
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine // 设置引擎
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix, // 前缀为传入对象的前缀
		parent: group,                 // 父分组为该对象
		engine: engine,                // 引擎为原引擎
	}
	engine.groups = append(engine.groups, newGroup) // 加入分组切片中
	return newGroup
}

// addRoute 添加路由方法，方法，pattern，handler函数
// 这里的pattern可以理解为路由路径
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 请求，会将路由和处理方法注册到映射表router中
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 请求
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath) // 绝对路径
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// check if file exits and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static serve static files，用户可用此方法把磁盘上的某个文件root映射到路由中
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root)) // 文件地址和路由
	urlPattern := path.Join(relativePath, "/*filepath")                // 地址参数
	// Register GET handlers
	group.GET(urlPattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// Run 运行，ListenAndServe 的包装
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
