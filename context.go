package gooo

import (
	"net/http"
	"time"
)

// H 自定义类型表示键值对数据
type H map[string]interface{}

// Context 上下文结构体
type Context struct {
	// 原始对象
	Writer   http.ResponseWriter
	Req      *http.Request
	Response *Response // 新增响应模块引用
	// 请求信息
	Path   string
	Method string
	Params map[string]string // 路由参数
	// 中间件数据
	keys     map[string]interface{}
	handlers []HandlerFunc // 中间件链
	index    int           // 当前执行的中间件索引
	aborted  bool          // 是否已终止
	engine   *Engine
}

// 构造函数
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:   w,
		Req:      req,
		Response: &Response{Writer: w}, // 初始化响应模块
		Path:     req.URL.Path,
		Method:   req.Method,
		Params:   make(map[string]string),
		keys:     make(map[string]interface{}),
	}
}

// Set 存储中间件数据
func (c *Context) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

// Get 获取中间件数据
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.keys[key]
	return
}

// MustGet 获取中间件数据，不存在则panic
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// 参数获取方法
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 获取路由参数
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// 获取请求参数(查询参数或表单参数)
func (c *Context) GetParam(key string) string {
	if c.Req == nil {
		return ""
	}
	switch c.Method {
	case http.MethodGet:
		return c.Query(key)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return c.PostForm(key)
	default:
		return ""
	}
}

// 设置响应头
func (c *Context) SetHeader(key, value string) {
	c.Response.SetHeader(key, value)
}

// 设置响应头
func (c *Context) SetContentType(value string) {
	c.Response.SetContentType(value)
}

// 设置响应状态码
func (c *Context) Status(code int) {
	c.Response.Status(code) // 双向同步
}

// 获取响应头
func (c *Context) GetContentType(name string) string {
	return c.Response.GetContentType(name)
}

// 获取响应状态码
func (c *Context) GetStatusCode() int {
	return c.Response.StatusCode
}

// 设置响应内容 JSON
func (c *Context) JSON(code int, obj interface{}) {
	c.Response.JSON(code, obj)
}

// 设置响应内容 String
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Response.String(code, format, values...)
}

// 设置响应内容 HTML
func (c *Context) HTML(code int, html string) {
	c.Response.HTML(code, html)
}

// 1. 添加请求上下文超时控制
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.Req != nil && c.Req.Context() != nil {
		return c.Req.Context().Deadline()
	}
	return
}

// 2. 统一参数获取方法
func (c *Context) GetParamWithDefault(key string, defaultValue string) string {
	if val := c.GetParam(key); val != "" {
		return val
	}
	return defaultValue
}

// Abort 终止后续中间件的执行
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted 检查是否已终止
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Next 执行中间件链中的下一个处理器
func (c *Context) Next() {
	for c.index < len(c.handlers)-1 && !c.aborted {
		c.index++
		handler := c.handlers[c.index]
		handler(c)
	}
}
