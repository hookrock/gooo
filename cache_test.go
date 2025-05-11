package gooo

import (
	"sync"
	"testing"
	"time"
)

// TestBasicCache 测试基础缓存操作（Set/Get/Del）
func TestBasicCache(t *testing.T) {
	cache := NewCache()

	// 设置缓存
	cache.SetCache("key1", "value1", 1*time.Hour)

	// 获取缓存
	val, ok := cache.GetCache("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// 删除缓存
	cache.DelCache("key1")
	_, ok = cache.GetCache("key1")
	if ok {
		t.Errorf("Expected key 'key1' to be deleted")
	}
}

// TestLRUEviction 测试 LRU 淘汰策略
func TestLRUEviction(t *testing.T) {
	cache := NewCache()
	cache.maxSize = 2 // 限制缓存容量为 2

	// 插入 3 个缓存项，触发 LRU 淘汰
	cache.SetCache("key1", "value1", 1*time.Hour)
	cache.SetCache("key2", "value2", 1*time.Hour)
	cache.SetCache("key3", "value3", 1*time.Hour)

	// 检查 key1 是否被淘汰
	val, ok := cache.GetCache("key1")
	if ok {
		t.Errorf("Expected key 'key1' to be evicted by LRU, but it's still present")
	}
	if val != nil {
		t.Errorf("Expected key 'key1' to be evicted by LRU, but it's still present")
	}

	// 检查 key2 和 key3 是否存在
	val, ok = cache.GetCache("key2")
	if !ok || val != "value2" {
		t.Errorf("Expected key 'key2' to exist")
	}

	val, ok = cache.GetCache("key3")
	if !ok || val != "value3" {
		t.Errorf("Expected key 'key3' to exist")
	}
}

// TestExpiration 测试过期缓存
func TestExpiration(t *testing.T) {
	cache := NewCache()

	// 设置一个 100ms 后过期的缓存
	cache.SetCache("key1", "value1", 100*time.Millisecond)

	// 立即获取，应该存在
	val, ok := cache.GetCache("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected key 'key1' to exist immediately")
	}

	// 等待 200ms，缓存过期
	time.Sleep(200 * time.Millisecond)

	// 再次获取，应该不存在
	_, ok = cache.GetCache("key1")
	if ok {
		t.Errorf("Expected key 'key1' to be expired")
	}
}

// TestGCMechanism 测试 GC 机制
func TestGCMechanism(t *testing.T) {
	cache := NewCache()

	// 启动 GC，间隔 100ms
	cache.StartGC(100 * time.Millisecond)

	// 设置一个 50ms 后过期的缓存
	cache.SetCache("key1", "value1", 50*time.Millisecond)

	// 等待 GC 执行
	time.Sleep(200 * time.Millisecond)

	// 检查缓存是否被 GC 清理
	_, ok := cache.GetCache("key1")
	if ok {
		t.Errorf("Expected key 'key1' to be cleaned up by GC")
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	cache := NewCache()

	const numGoroutines = 100
	const numKeys = 10

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key" + string(rune('a'+idx%numKeys))
			cache.SetCache(key, idx, 1*time.Hour)
			val, ok := cache.GetCache(key)
			if !ok || val != idx {
				t.Errorf("Concurrent access failed for key %s", key)
			}
		}(i)
	}

	wg.Wait()
}

// TestGetCacheMulti 测试批量获取缓存
func TestGetCacheMulti(t *testing.T) {
	cache := NewCache()

	// 设置多个缓存
	cache.SetCache("key1", "value1", 1*time.Hour)
	cache.SetCache("key2", "value2", 1*time.Hour)
	cache.SetCache("key3", "value3", 1*time.Hour)

	// 获取多个缓存
	keys := []string{"key1", "key2", "key3"}
	result := cache.GetCacheMulti(keys)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	if result["key1"] != "value1" || result["key2"] != "value2" || result["key3"] != "value3" {
		t.Errorf("Unexpected values in GetCacheMulti result")
	}
}

// TestGetCacheAll 测试获取所有缓存
func TestGetCacheAll(t *testing.T) {
	cache := NewCache()

	// 设置多个缓存
	cache.SetCache("key1", "value1", 1*time.Hour)
	cache.SetCache("key2", "value2", 1*time.Hour)
	cache.SetCache("key3", "value3", 1*time.Hour)

	// 获取所有缓存
	result := cache.GetCacheAll()

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	if result["key1"] != "value1" || result["key2"] != "value2" || result["key3"] != "value3" {
		t.Errorf("Unexpected values in GetCacheAll result")
	}
}
