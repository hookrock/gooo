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
	t.Skip("TODO: Implement parameterized routes")

	engine := New()
	engine.GET("/user/:name", func(c *Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello "+name)
	})

	t.Run("route with params", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/user/john", nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if body := w.Body.String(); body != "Hello john" {
			t.Errorf("Expected body 'Hello john', got '%s'", body)
		}
	})
}
