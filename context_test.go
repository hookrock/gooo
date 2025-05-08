package gooo

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestContext_RequestResponse(t *testing.T) {
	// 模拟请求
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/test",
		},
	}
	w := httptest.NewRecorder()

	// 创建Context
	c := &Context{
		Writer: w,
		Req:    req,
	}

	// 测试String方法
	c.String(http.StatusOK, "Hello %s", "World")

	// 验证响应
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("Expected body 'Hello World', got '%s'", w.Body.String())
	}
}

func TestContext_QueryParams(t *testing.T) {
	// 模拟带查询参数的请求
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "name=test&id=123",
		},
	}
	c := &Context{
		Req: req,
	}

	// 测试Query方法
	if name := c.Query("name"); name != "test" {
		t.Errorf("Expected query name 'test', got '%s'", name)
	}
	if id := c.Query("id"); id != "123" {
		t.Errorf("Expected query id '123', got '%s'", id)
	}
	if missing := c.Query("missing"); missing != "" {
		t.Errorf("Expected empty query for missing param, got '%s'", missing)
	}
}

func TestContext_PostForm(t *testing.T) {
	// 模拟POST表单请求
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"Content-Type": []string{"application/x-www-form-urlencoded"},
		},
		PostForm: url.Values{
			"username": []string{"admin"},
			"password": []string{"secret"},
		},
	}
	c := &Context{
		Req: req,
	}

	// 测试PostForm方法
	if username := c.PostForm("username"); username != "admin" {
		t.Errorf("Expected form username 'admin', got '%s'", username)
	}
	if password := c.PostForm("password"); password != "secret" {
		t.Errorf("Expected form password 'secret', got '%s'", password)
	}
}

func TestContext_Status(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{
		Writer: w,
	}

	// 测试状态码设置
	c.Status(http.StatusNotFound)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
