package main

import (
	"fmt"
	"gCache/gcache"
	"log"
	"net/http"
)

func createGroup() *gcache.Group {
	return gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务
func startCacheServer(addr string, addrs []string, g *gcache.Group) {
	peers := gcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("gcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
