package gooo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response 封装所有响应逻辑
type Response struct {
	Writer     http.ResponseWriter
	StatusCode int
}

// 基础方法
func (r *Response) SetHeader(key, value string) {
	r.Writer.Header().Set(key, value)
}

func (r *Response) SetContentType(value string) {
	r.SetHeader("Content-Type", value)
}

func (r *Response) Status(code int) {
	r.StatusCode = code
	r.Writer.WriteHeader(code)
}

func (r *Response) JSON(code int, obj interface{}) {
	r.SetContentType("application/json")
	r.Status(code)
	if err := json.NewEncoder(r.Writer).Encode(obj); err != nil {
		http.Error(r.Writer, err.Error(), 500)
	}
}

func (r *Response) HTML(code int, html string) {
	r.SetContentType("text/html")
	r.Status(code)
	r.Writer.Write([]byte(html))
}

// response.go 需新增的方法
func (r *Response) String(code int, format string, values ...any) {
	r.SetContentType("text/plain")
	r.Status(code)
	r.Writer.Write(fmt.Appendf(nil, format, values...))
}

func (r *Response) Data(code int, data []byte) {
	r.Status(code)
	r.Writer.Write(data)
}

func (r *Response) Fail(code int, err string) {
	r.JSON(code, H{"message": err})
}

func (r *Response) GetContentType(name string) string {
	if name == "" {
		name = "Content-Type"
	}
	return r.Writer.Header().Get(name)
}

func (r *Response) InjectDebugHTML(msg string) {
	if r.GetContentType("Content-Type") == "text/html" {
		debugHTML := fmt.Sprintf(`<div>%s</div>`, msg)
		r.Writer.Write([]byte(debugHTML))
	}
}

// 建议新增以下方法：
func (r *Response) Redirect(code int, location string) {
	r.SetHeader("Location", location)
	r.Status(code)
}

func (r *Response) Attachment(filename string) {
	r.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
}

// 错误处理优化
func (r *Response) Error(code int, err error) {
	r.JSON(code, H{"error": err.Error()})
}
