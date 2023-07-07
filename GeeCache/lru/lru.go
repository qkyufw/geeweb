package lru

import "container/list"

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes int64                    // cache允许的最大内存
	nbytes   int64                    // cache当前已使用的内存
	ll       *list.List               // go 标准库实现的双向链表，存放所有的值
	cache    map[string]*list.Element // 字典，键为字符串，值为链表中对应的节点
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调记录，可以为nil
}

// entry 双向链表结点的数据类型，
// 在链表中仍保存每个键值对应的key，
// 删除队首节点时，需要使用key从字典删除对应的映射
type entry struct {
	key   string
	value Value
}

// New is the Constructor of Cache，用于实例化
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,   // 最大内存数设置为传入的数字
		ll:        list.New(), // 新建一个双向链表
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted, // 回调记录
	}
}

// Add 新增与修改
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 如果键值存在
		c.ll.MoveToFront(ele)    // 更新并移动到队尾
		kv := ele.Value.(*entry) // 更新缓存大小
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { // 不存在，是新增场景，先添加节点
		ele := c.ll.PushFront(&entry{key, value})        // 添加节点到头部
		c.cache[key] = ele                               // 再向字典中添加映射关系
		c.nbytes += int64(len(key)) + int64(value.Len()) // 更新缓存大小
	}
	// 如果 c.nbytes超过了最大值，就移除最少访问节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get look ups a key's value，查找功能
// 节点背找到，则被使用，移动到最后去
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { // 从字典中找到对应的双向链表的节点
		c.ll.MoveToFront(ele) // 把该节点移动到队尾
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes teh oldest item
// 只是用于缓存淘汰，移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 取出到队首节点，从链表中删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                // 从字典中c.cache 删除该节点的映射关系
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新当前所用的内存 c.nbytes
		if c.OnEvicted != nil {                                // 如果回调函数不为nil就调用
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len 用于获取添加了多少条数据
func (c *Cache) Len() int {
	return c.ll.Len()
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int // 用于返回值所占用的内存大小
}
