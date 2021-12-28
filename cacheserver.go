package main

import (
	"fmt"
	"gcache/cache"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Info("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务
func startCacheServer(addr string, addrs []string, g *cache.Group) {
	peers := cache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Infof("cache is running at:%s", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
