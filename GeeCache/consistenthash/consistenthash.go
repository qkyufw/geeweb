package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 解决分布式访问谁的问题
// 对于给定key，确保每次都能选择同一个节点，使用了hash算法
// 节点数量也可能发生变化，一致性hash解决变化的情况

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
// 主数据结构，4个成员变量
type Map struct {
	hash     Hash           // Hash 函数
	replicas int            // 虚拟节点倍数
	keys     []int          // Sorted，哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil { // Hash 可以采用注入的方式，也能替换成自己的
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) { // 允许传入0个或多个真实节点的名称
	for _, key := range keys { // 对每个真实节点对应创建m.replicas个虚节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // m.hash计算虚拟节点的哈希值
			m.keys = append(m.keys, hash)                      // 添加到环上
			m.hashMap[hash] = key                              // 增加映射关系
		}
	}
	sort.Ints(m.keys) // 环上的哈希值排序 todo why可以这样排序
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key))) // 计算节点
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { // 顺时针找到第一个匹配的虚拟节点下标idx，获取对应hash值
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]] // 通过hashMap映射得到真实节点
}
