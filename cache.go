package gooo

import (
	"container/list"
	"sync"
	"time"
)

type CacheItem struct {
	key        string
	value      any
	expiration int64 // 过期时间（纳秒）
}

type Cache struct {
	maxSize int
	list    *list.List
	items   map[string]*list.Element
	mu      sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		maxSize: 10000, // 默认缓存最大容量为10000
		list:    list.New(),
		items:   make(map[string]*list.Element),
	}
}

// SetCache 设置缓存
func (c *Cache) SetCache(key string, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expiration := time.Now().Add(duration).UnixNano()

	// 如果已存在，先删除旧值
	if el, ok := c.items[key]; ok {
		c.list.Remove(el)
	}
	// 插入新值到链表头
	newItem := &CacheItem{
		key:        key,
		value:      value,
		expiration: expiration,
	}
	el := c.list.PushFront(newItem)
	c.items[key] = el

	if c.list.Len() > c.maxSize {
		c.removeOldest()
	}
}

// 删除最久未使用的项
func (c *Cache) removeOldest() {
	if el := c.list.Back(); el != nil {
		c.list.Remove(el)
		item := el.Value.(*CacheItem)
		delete(c.items, item.key)
	}
}

// Delete 删除缓存
func (c *Cache) DelCache(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.list.Remove(el)
		delete(c.items, key)
	}
}

func (c *Cache) StartGC(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			c.gc()
		}
	}()
}

func (c *Cache) gc() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, el := range c.items {
		item := el.Value.(*CacheItem)
		if item.expiration < now {
			c.list.Remove(el)
			delete(c.items, key)
		}
	}
}

// GetCache 获取缓存值
func (c *Cache) GetCache(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	cacheItem := item.Value.(*CacheItem)

	if cacheItem.expiration < time.Now().UnixNano() {
		return nil, false
	}

	return cacheItem.value, true
}

// GetCacheMulti 获取多个缓存
func (c *Cache) GetCacheMulti(keys []string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]any)
	now := time.Now().UnixNano()

	for _, key := range keys {
		item, found := c.items[key]
		if !found {
			continue
		}

		cacheItem := item.Value.(*CacheItem)
		if cacheItem.expiration < now {
			c.list.Remove(item)
			delete(c.items, key)
			continue
		}

		// 可选：更新访问顺序（LRU 行为）
		c.list.MoveToFront(item)

		result[key] = cacheItem.value
	}

	return result
}

// GetCacheAll 获取所有缓存
func (c *Cache) GetCacheAll() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]any)
	now := time.Now().UnixNano()
	count := 0

	for e := c.list.Front(); e != nil; {
		count++
		cacheItem := e.Value.(*CacheItem)
		next := e.Next() // 提前获取下一个节点

		if cacheItem.expiration < now {
			c.list.Remove(e)
			delete(c.items, cacheItem.key)
			e = next
			continue
		}

		c.list.MoveToFront(e)
		result[cacheItem.key] = cacheItem.value
		e = next
	}

	return result
}

// GetCacheCount 获取缓存个数
func (c *Cache) GetCacheCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.list.Len()
}
