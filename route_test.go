package gooo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine_GET(t *testing.T) {
	engine := New()
	engine.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "GET test")
	})

	t.Run("valid GET request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if body := w.Body.String(); body != "GET test" {
			t.Errorf("Expected body 'GET test', got '%s'", body)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/notfound", nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestEngine_POST(t *testing.T) {
	engine := New()
	engine.POST("/submit", func(c *Context) {
		c.String(http.StatusOK, "POST received")
	})

	t.Run("valid POST request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/submit", nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if body := w.Body.String(); body != "POST received" {
			t.Errorf("Expected body 'POST received', got '%s'", body)
		}
	})
}

func TestEngine_Params(t *testing.T) {
	engine := New()

	// 单参数
	engine.GET("/user/:name", func(c *Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello "+name)
	})

	// 多参数
	engine.GET("/:ok/:ada", func(c *Context) {
		ok := c.Param("ok")
		ada := c.Param("ada")
		c.String(http.StatusOK, "ok: "+ok+", ada: "+ada)
	})

	// 混合路径
	engine.GET("/:dad/o/jj/:fa", func(c *Context) {
		dad := c.Param("dad")
		fa := c.Param("fa")
		c.String(http.StatusOK, "dad: "+dad+", fa: "+fa)
	})

	// 通配符
	engine.GET("/:fafa/*tsfs", func(c *Context) {
		fafa := c.Param("fafa")
		tsfs := c.Param("tsfs")
		c.String(http.StatusOK, "fafa: "+fafa+", tsfs: "+tsfs)
	})

	t.Run("single param", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/user/john", nil)
		engine.ServeHTTP(w, req)

		if body := w.Body.String(); body != "Hello john" {
			t.Errorf("Expected 'Hello john', got '%s'", body)
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/alice/bob", nil)
		engine.ServeHTTP(w, req)

		if body := w.Body.String(); body != "ok: alice, ada: bob" {
			t.Errorf("Expected 'ok: alice, ada: bob', got '%s'", body)
		}
	})

	t.Run("mixed path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/mum/o/jj/dad", nil)
		engine.ServeHTTP(w, req)

		if body := w.Body.String(); body != "dad: mum, fa: dad" {
			t.Errorf("Expected 'dad: mum, fa: dad', got '%s'", body)
		}
	})

	t.Run("wildcard path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello/world/this/is/a/test", nil)
		engine.ServeHTTP(w, req)

		if body := w.Body.String(); body != "fafa: hello, tsfs: world/this/is/a/test" {
			t.Errorf("Expected 'fafa: hello, tsfs: world/this/is/a/test', got '%s'", body)
		}
	})
}
