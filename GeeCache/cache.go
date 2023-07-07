package GeeCache

import (
	"GeeCache/lru"
	"sync"
)

// 实例化lru，封装get 和 add方法，添加互斥锁mu

type cache struct {
	mu         sync.Mutex // 添加并发特性
	lru        *lru.Cache
	cacheBytes int64 // 缓存允许的最大内存大小
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟初始化，提高性能，减少程序内存需求
	if c.lru == nil { // 判断lru是否为空，为空就创建实例
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value) // 不为空就添加进去
}

// 从缓存中获取指定键的缓存条目。
// 同样通过互斥对共享资源进行保护
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 判断lru是否为空
		return // 为空直接返回
	}

	// 不为空，调用Get来获取指定键的值
	// Get方法返回一个接口类型的值
	// 在返回时需要做一个断言操作，转化为ByteView
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
