package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	element, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) Add(key string, value Value) {
	element, ok := c.cache[key]
	if ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len()) // 键不变，值更新的情况下，更新旧值的大小
		kv.value = value                                          // 更新值
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		kv := element.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
	}
}
