package gooo

import (
	"net/http"
	"strings"
)

// 定义一个Hanlder 自定义处理函数类型
type HandlerFunc func(c *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}

	return engine
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)

	// 1. 收集所有中间件(全局+匹配路由组)
	var handlers []HandlerFunc

	// 添加全局中间件
	handlers = append(handlers, engine.RouterGroup.middlewares...)

	// 添加匹配路由组的中间件
	for _, group := range engine.groups {
		if group != engine.RouterGroup && strings.HasPrefix(r.URL.Path, group.prefix) {
			handlers = append(handlers, group.middlewares...)
		}
	}

	// 2. 添加路由处理器
	handlers = append(handlers, func(c *Context) {
		engine.router.handler(c)
	})

	// 3. 执行处理链
	c.handlers = handlers
	c.index = -1
	c.Next()
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) Run(addr ...string) (err error) {
	port := resolveAddress(addr)
	return http.ListenAndServe(port, engine)
}
