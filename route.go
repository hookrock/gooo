package gooo

import (
	"fmt"
	"net/http"
	"strings"
)

// RouterGroup 路由组结构体
type RouterGroup struct {
	prefix string       // 路由组前缀
	middlewares []HandlerFunc // 中间件列表
	parent *RouterGroup // 父路由组
	engine *Engine      // 所有路由组共享一个Engine实例
}

// Use 注册中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	if group.middlewares == nil {
		group.middlewares = make([]HandlerFunc, 0)
	}
	for _, m := range middlewares {
		if m != nil {
			group.middlewares = append(group.middlewares, m)
		}
	}
	// 确保中间件被添加到engine的groups列表中
	if group.engine != nil {
		group.engine.groups = append(group.engine.groups, group)
	}
}

// Group 创建新的路由组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	return &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: group.engine,
	}
}

// GET 添加GET路由
func (group *RouterGroup) GET(path string, handler HandlerFunc) {
	fullPath := group.prefix + path
	if fullPath[0] != '/' {
		fullPath = "/" + fullPath
	}
	group.engine.router.addRoute("GET", fullPath, handler)
}

// POST 添加POST路由
func (group *RouterGroup) POST(path string, handler HandlerFunc) {
	fullPath := group.prefix + path
	if fullPath[0] != '/' {
		fullPath = "/" + fullPath
	}
	group.engine.router.addRoute("POST", fullPath, handler)
}

type router struct {
	handlers map[string]HandlerFunc
	roots    map[string]*trie
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandlerFunc),
		roots:    make(map[string]*trie),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, path string, handler HandlerFunc) {
	parts := parsePattern(path)

	// 检查路由冲突
	if root, ok := r.roots[method]; ok {
		if conflictNode := root.search(parts); conflictNode != nil {
			panic(fmt.Sprintf("路由冲突: %s %s 与 %s %s", method, path, method, conflictNode.part))
		}
	} else {
		r.roots[method] = &trie{}
	}

	r.roots[method].insert(path, parts)
	key := method + "-" + path
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*trie, map[string]string) {
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	parts := parsePattern(path)
	node := root.search(parts)
	if node == nil {
		return nil, nil
	}

	params := make(map[string]string)
	matchedParts := parsePattern(node.part)
	for i, part := range matchedParts {
		if part[0] == ':' {
			params[part[1:]] = parts[i]
		}
		if part[0] == '*' && len(part) > 1 {
			params[part[1:]] = strings.Join(parts[i:], "/")
			break
		}
	}
	return node, params
}

func (r *router) handler(c *Context) {
	node, params := r.getRoute(c.Method, c.Path)
	if node != nil {
		key := c.Method + "-" + node.pattern
		c.Params = params
		if handler, ok := r.handlers[key]; ok {
			handler(c)
			return
		}
	}
	c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
}
