package gooo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestContext_RequestResponse(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Path: "/test",
		},
	}
	w := httptest.NewRecorder()

	c := &Context{
		Writer: w,
		Req:    req,
		Response: &Response{
			Writer: w,
		},
	}

	// Test String response
	c.String(http.StatusOK, "Hello %s", "World")
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("Expected body 'Hello World', got '%s'", w.Body.String())
	}
}

func TestContext_QueryParams(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			RawQuery: "name=test&id=123",
		},
	}
	c := &Context{Req: req}

	if name := c.Query("name"); name != "test" {
		t.Errorf("Expected query name 'test', got '%s'", name)
	}
	if id := c.Query("id"); id != "123" {
		t.Errorf("Expected query id '123', got '%s'", id)
	}
}

func TestContext_PostForm(t *testing.T) {
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
	c := &Context{Req: req}

	if username := c.PostForm("username"); username != "admin" {
		t.Errorf("Expected form username 'admin', got '%s'", username)
	}
}

func TestContext_Status(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Writer: w, Response: &Response{Writer: w}}

	c.Status(http.StatusNotFound)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestContext_JSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Writer: w, Response: &Response{Writer: w}}

	data := map[string]string{"key": "value"}
	c.JSON(http.StatusOK, data)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type application/json")
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("Expected JSON value 'value', got '%s'", result["key"])
	}
}

func TestContext_MiddlewareChain(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{
		engine:   New(),
		index:    -1, // 显式设置初始索引
		Writer:   w,
		handlers: make([]HandlerFunc, 0),
		Response: &Response{Writer: w},
	}

	// 使用标准中间件注册方式
	c.engine.Use(func(c *Context) {
		c.Set("mw1", "executed")
		c.Next() // 必须调用Next继续链式调用
	})
	c.engine.Use(func(c *Context) {
		c.Set("mw2", "executed")
	})

	// 模拟完整请求处理流程
	c.engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	// Add test middlewares
	c.handlers = append(c.handlers, func(c *Context) {
		c.Set("mw1", "executed")
		c.Next()
	})
	c.handlers = append(c.handlers, func(c *Context) {
		c.Set("mw2", "executed")
	})

	c.Next()

	if val, _ := c.Get("mw1"); val != "executed" {
		t.Error("Middleware 1 not executed")
	}

	if val, _ := c.Get("mw2"); val != "executed" {
		t.Error("Middleware 2 not executed")
	}
}

func TestContext_Abort(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{
		engine:   New(),
		index:    -1, // 显式设置初始索引
		Writer:   w,
		handlers: make([]HandlerFunc, 0),
		Response: &Response{Writer: w},
	}

	c.handlers = append(c.handlers, func(c *Context) {
		c.Abort()
	})
	c.handlers = append(c.handlers, func(c *Context) {
		t.Error("This middleware should not be executed")
	})

	c.Next()

	if !c.IsAborted() {
		t.Error("Context should be aborted")
	}
}

func TestContext_Deadline(t *testing.T) {
	c := &Context{Req: &http.Request{}, Response: &Response{Writer: httptest.NewRecorder()}}
	_, ok := c.Deadline()
	if ok {
		t.Error("Expected no deadline by default")
	}
}

func TestContext_GetParamWithDefault(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			RawQuery: "name=test",
		},
	}
	c := &Context{Req: req,
		Method: req.Method, Response: &Response{Writer: httptest.NewRecorder()}}

	if val := c.GetParamWithDefault("name", "default"); val != "test" {
		t.Errorf("Should return query param value: %s", c.Query("name"))
	}
	if val := c.GetParamWithDefault("missing", "default"); val != "default" {
		t.Errorf("Should return default value for missing param")
	}
}

func TestContext_SessionOperations(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{
		Writer: w,
		Req:    &http.Request{},
		engine: &Engine{
			sessionManager: NewSessionManager(
				NewMemoryStore(time.Minute),
			),
		},
		Response: &Response{Writer: w},
	}

	// 显式初始化 CookieOpts（可选）
	c.engine.sessionManager.CookieOpts = CookieConfig{
		Secure:   false,
		SameSite: SameSiteMode,
		MaxAge:   int(DefaultExpire.Seconds()),
	}

	// 测试会话创建
	c.StartSession()
	if c.Session == nil {
		t.Error("会话应被初始化")
	}

	// 测试会话续期
	oldID := c.SessionID
	c.RenewSession()
	if c.SessionID == oldID {
		t.Errorf("会话ID应在续期后改变，旧ID: %s, 新ID: %s", oldID, c.SessionID)
	}

	// 测试会话销毁
	c.DestroySession()
	if c.Session != nil {
		t.Error("会话应被销毁")
	}
}
