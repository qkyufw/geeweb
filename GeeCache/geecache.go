package GeeCache

import (
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key
// Getter接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function
// 回调函数GetterFunc，实现Getter接口中的方法
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

// A Group is a cache namespace and associated data loaded spread over
// 可以认为是一个缓存的命名空间
type Group struct {
	name      string // 每个Group有唯一的名称name
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 一开始实现的并发缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group，实例化Group
// 将Group存储在全局变量groups中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup
// or nil if there's no such group
func GetGroup(name string) *Group {
	mu.RLock() // 使用了只读锁，不涉及任何冲突变量的写操作
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 从mainCache中查找缓存，如果存在就返回缓存值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 不存在就调用load方法
	return g.load(key)
}

// load 继续调用getLocally方法（分布式场景会继续调用getFromPeer从其他节点获取）
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// getLocally 调用用户回调函数g.getter.Get()获取元数据，并将数据添加到缓存mainCache中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 添加数据到缓存中去
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
