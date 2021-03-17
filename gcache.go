package gCache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64) *Group {
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("gcache hit")
		return v, nil
	}

	// TODO load
	return ByteView{}, fmt.Errorf("TODO")
}
