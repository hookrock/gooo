package gooo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareExecution(t *testing.T) {
	engine := New()
	executionLog := make([]string, 0)

	// 全局中间件1
	engine.Use(func(c *Context) {
		executionLog = append(executionLog, "global1-start")
		c.Next()
		executionLog = append(executionLog, "global1-end")
	})

	// 全局中间件2
	engine.Use(func(c *Context) {
		executionLog = append(executionLog, "global2-start")
		c.Next()
		executionLog = append(executionLog, "global2-end")
	})

	// 路由组中间件
	api := engine.Group("/api")
	api.Use(func(c *Context) {
		executionLog = append(executionLog, "group1-start")
		c.Next()
		executionLog = append(executionLog, "group1-end")
	})

	// 路由处理器
	api.GET("/test", func(c *Context) {
		executionLog = append(executionLog, "handler")
		c.String(http.StatusOK, "OK")
	})

	// 发送测试请求
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// 验证关键执行点
	checkExecution := func(phase string) bool {
		for _, step := range executionLog {
			if step == phase {
				return true
			}
		}
		return false
	}

	// 验证关键中间件执行情况
	requiredSteps := []string{
		"global1-start", "handler",
	}

	for _, step := range requiredSteps {
		if !checkExecution(step) {
			t.Errorf("missing required execution step: %s", step)
		}
	}
}

func TestMiddlewareAbort(t *testing.T) {
	engine := New()
	executed := false

	// 中间件提前终止
	engine.Use(func(c *Context) {
		c.String(http.StatusForbidden, "Access denied")
		c.Abort()
	})

	// 这个处理器不应该执行
	engine.GET("/", func(c *Context) {
		executed = true
	})

	// 发送测试请求
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// 验证处理器未执行
	if executed {
		t.Error("handler should not be executed after abort")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}
