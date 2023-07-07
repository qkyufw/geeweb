package singleflight

import "sync"

// call 正在进行中或已结束的请求
type call struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// Group single flight 的数据主结构，管理不同key的请求call
type Group struct {
	mu sync.Mutex // protects m 保护成员变量m不被并发读写而加上的锁
	m  map[string]*call
}

// Do 针对相同的key，无论Do被调用多少次，fn都只会被调用一次
// 等待fn调用结束后，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil { // 延迟初始化
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  // 请求发起前加锁
	g.m[key] = c // 添加到g.m，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新g.m
	g.mu.Unlock()

	return c.val, c.err
}
