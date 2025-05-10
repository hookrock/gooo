package gooo

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// H 自定义类型表示键值对数据
type H map[string]interface{}

// Context 上下文结构体
type Context struct {
	// 原始对象
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	Params map[string]string // 路由参数
	// 中间件数据
	keys     map[string]interface{}
	handlers []HandlerFunc // 中间件链
	index    int           // 当前执行的中间件索引
	aborted  bool          // 是否已终止
	// 响应信息
	StatusCode int

	engine *Engine
}

// 构造函数
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		Params: make(map[string]string),
		keys:   make(map[string]interface{}),
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
	c.Writer.Header().Set(key, value)
}

// 设置响应头
func (c *Context) SetContentType(value string) {
	c.SetHeader("Content-Type", value)
}

// 获取响应头
func (c *Context) GetContentType(name string) string {
	if name == "" {
		name = "Content-Type"
	}
	return c.Writer.Header().Get(name)
}

// 设置响应状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// 获取响应状态码
func (c *Context) GetStatusCode() int {
	return c.StatusCode
}

// 设置响应内容 String
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetContentType("text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// 设置响应内容 Json
func (c *Context) JSON(code int, obj interface{}) {
	c.SetContentType("application/json")
	c.Status(code)
	enc := json.NewEncoder(c.Writer)
	if err := enc.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// // 设置响应内容 XML
// func (c *Context) XML(code int, obj interface{}) {
// 	c.SetContentType("application/xml")
// 	c.Status(code)
// 	enc := xml.NewEncoder(c.Writer)
// 	if err := enc.Encode(obj); err != nil {
// 		http.Error(c.Writer, err.Error(), 500)
// 	}
// }

// 设置响应内容 Data
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// 设置响应内容 HTML
func (c *Context) HTML(code int, html string) {
	c.SetContentType("text/html")
	c.Status(code)
	if _, err := c.Writer.Write([]byte(html)); err != nil {
		log.Printf("HTML写入失败: %v", err)
	}
}

// 设置调试信息
func (c *Context) debugInjectHTML(msg string) {
	// 检查响应是否为 HTML 类型（可以根据实际需求扩展）
	if c.GetContentType("Content-Type") == "text/html" {
		// 构造调试面板的 HTML 内容
		debugHTML := fmt.Sprintf(`
			<div style="position:fixed;bottom:0;left:0;width:100%%;background:#000;color:#fff;font-family:monospace;padding:10px;z-index:9999;">
				<pre>%s</pre>
			</div>
		`, msg)

		// 修改原始 HTML 内容（假设你已经缓存了原始响应）
		c.Writer.Write([]byte(debugHTML))
	}
}

// 在调试模式下向网页端输出调试信息
// func (c *Context) debugPrint(format string, values ...interface{}) {
// 	if os.Getenv("DEBUG") != "" || IsDebugMode() {
// 		// 2. 构造调试信息
// 		debugMsg := "[DEBUG] " + fmt.Sprintf(format, values...)

// 		// 3. 输出到日志
// 		log.Print(debugMsg)

// 		// 4. 输出到网页端（仅在 HTML 响应中）
// 		c.debugInjectHTML(debugMsg)
// 	}
// }

// Fail 设置错误信息
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
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
