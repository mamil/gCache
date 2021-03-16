package lru

import (
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0))
	lru.Add("key1", String("123"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "123" {
		t.Fatalf("cache hit key1=1234 failed")
	} else if ok {
		t.Log(string(v.(String)))
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}
