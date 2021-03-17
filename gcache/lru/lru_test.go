package lru

import (
	"fmt"
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
		t.Fatalf("cache hit key1=123 failed")
	} else if ok {
		t.Log(string(v.(String)))
		s := "v:" + fmt.Sprint(v.Len()) + ", k:" + fmt.Sprint(len("key1"))
		t.Log(s)
		t.Log(lru.ShowSize())
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestGet2(t *testing.T) {
	lru := New(int64(6))
	lru.Add("key1", String("123"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "123" {
		t.Log("cache deleted")
		t.Log(lru.ShowSize())
	}

}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap))
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
