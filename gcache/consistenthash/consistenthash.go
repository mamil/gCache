package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32 // hash算法，默认用crc32.ChecksumIEEE，也可以替换成自己的

type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射,键是虚拟节点的哈希值，值是真实节点的名称。
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 这里只保存了key，用来指明传入的key应该对应哪个真实的节点值，这样找到节点之后，从那个节点去拿数据。并非真正存储数据
// 增加节点，传入真实节点，会创建m.replicas个虚拟节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 用编码来区别虚拟节点
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key // 增加虚拟节点和真实节点的映射
		}
	}
	sort.Ints(m.keys) // 环上的哈希值排序
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	ids := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[ids%len(m.keys)]]
}
