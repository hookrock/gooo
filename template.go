// 文件路径：gooo/template.go
package gooo

import (
	"html/template"
	"net/http"
	"strings"
)

// 模板渲染和静态文件配置



// Static 静态文件服务（原Engine方法）
func (engine *Engine) Static(relativePath, root string) {
	// 1. 添加路径规范化
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}
	handler := http.StripPrefix(relativePath, http.FileServer(http.Dir(root)))
	urlPattern := relativePath + "/*filepath"
	engine.GET(urlPattern, func(c *Context) {
		handler.ServeHTTP(c.Writer, c.Req)
	})
}

// View 渲染模板
func (c *Context) View(name string, data interface{}) {

	// 加载所有模板文件
	tmpl, err := template.ParseGlob("web/templates/*.tmpl")
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}

	// 执行模板
	if err := tmpl.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
		return
	}
}
