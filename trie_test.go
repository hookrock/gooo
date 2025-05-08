package gooo

import (
	"testing"
)

func TestTrieInsertAndSearch(t *testing.T) {
	root := &trie{
		part:     "/",
		children: make([]*trie, 0),
	}

	// 测试基本路由
	root.insert("/hello", parsePattern("/hello"))
	if node := root.search(parsePattern("/hello")); node == nil || node.pattern != "/hello" {
		t.Fatalf("basic route search failed")
	}

	// 测试不存在的路由
	if node := root.search(parsePattern("/notfound")); node != nil {
		t.Fatalf("should not find non-existent route")
	}
}

func TestTrieDynamicParam(t *testing.T) {
	root := &trie{}

	// 测试动态参数路由
	root.insert("/user/:id", parsePattern("/user/:id"))
	node := root.search(parsePattern("/user/123"))
	if node == nil || node.pattern != "/user/:id" {
		t.Fatalf("dynamic param route search failed")
	}
}

func TestTrieWildcard(t *testing.T) {
	root := &trie{}

	// 测试通配符路由
	root.insert("/static/*filepath", parsePattern("/static/*filepath"))
	node := root.search(parsePattern("/static/css/style.css"))
	if node == nil || node.pattern != "/static/*filepath" {
		t.Fatalf("wildcard route search failed")
	}
}

func TestTrieConflict(t *testing.T) {
	root := &trie{
		part:     "/",
		children: make([]*trie, 0),
	}

	// 测试路由共存 - 允许具体路径和参数路径共存
	root.insert("/user/name", parsePattern("/user/name"))
	root.insert("/user/:id", parsePattern("/user/:id")) // 这应该不会触发panic

	// 测试真正的冲突场景 - 相同路径不同处理
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("should panic on real route conflict")
		}
	}()
	root.insert("/user/name", parsePattern("/user/name"))
	root.insert("/user/name", parsePattern("/user/name")) // 这会触发panic
}
