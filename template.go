// 文件路径：gooo/template.go
package gooo

import (
	"html/template"
	"net/http"
	"os"
	"strings"
)

type TemplateEngine struct {
	templates *template.Template
	funcMap   template.FuncMap
}

type Config struct {
	Root string
	// 静态文件目录
	StaticPath string
	// 模板文件目录
	TemplatePath string
	// 后缀名
	Extension string
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: template.New(""),
		funcMap:   make(template.FuncMap),
	}
}

func (e *TemplateEngine) AddFunc(name string, fn interface{}) {
	if e.funcMap == nil {
		e.funcMap = make(template.FuncMap)
	}
	e.funcMap[name] = fn
}

func (e *TemplateEngine) Load(pattern string) error {
	var err error
	e.templates, err = template.New("").Funcs(e.funcMap).ParseGlob(pattern)
	return err
}

// Static 静态文件服务（原Engine方法）
func (e *Engine) Static(relativePath, root string) {
	// 1. 添加路径规范化
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		panic("Static file root directory does not exist: " + root)
	}
	handler := http.StripPrefix(relativePath, http.FileServer(http.Dir(root)))
	urlPattern := relativePath + "/*filepath"
	e.GET(urlPattern, func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Req)
	})
}

// View 渲染模板
func (c *Context) View(name string, data interface{}) {

	if c.engine.template.templates == nil {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	// 执行模板
	if err := c.engine.template.templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
		return
	}
}
