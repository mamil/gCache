package main

import (
	"container/list"
	"fmt"
)

type Cache struct {
	maxBytes  int64
	usedBytes int64
	ll      *list.List
	cache     map[string]*list.Element
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64) *Cache{
	return &Cache{
		maxBytes: maxBytes,
		ll: list.New(),
		cache: make(map[string]*list.Element)
	}
}

func (c *Cache) Get(key string) (value Value, ok bool){
	element, ok := c.cache[key]
	if ok{
		c.ll.MoveToFront(element)
		v := element.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func main() {
	var c Cache
	c.maxBytes = 64

	lru := New(int64(0))
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}
