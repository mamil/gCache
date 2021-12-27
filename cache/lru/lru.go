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
		kv := element.Value.(*entry)                              //从list中强制转回*entry 指针
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len()) // 键不变，值更新的情况下，更新旧值的大小
		kv.value = value                                          // 更新值
	} else {
		ele := c.ll.PushFront(&entry{key, value})           // 具体值存在list
		c.cache[key] = ele                                  // map里面存指向值的指针
		c.usedBytes += int64(len(key)) + int64(value.Len()) //用掉的空间包括键和值
	}

	for c.maxBytes != 0 && c.maxBytes < c.usedBytes { // 如果加入一个特别大的值，就要一直移除，直到不超限制
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

func (c *Cache) ShowSize() (int64, int64) {
	return c.maxBytes, c.usedBytes
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
