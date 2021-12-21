package main

import (
	"fmt"
	"gCache/gcache"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LeaderFunc func()

// 启动对外API服务
func startAPIServer(apiAddr string, g *gcache.Group) LeaderFunc {
	return func() {
		http.Handle("/api", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Query().Get("key")
				view, err := g.Get(key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(view.ByteSlice())

			}))
		log.Println("fontend server is running at", apiAddr)
		log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
	}
}

func startApiNode(apiAddr string, g *gcache.Group) {
	// 获取leader的回调函数
	leadFunc := startAPIServer(apiAddr, g)

	countStr := viper.GetString("APINODE.Count")
	log.Infof("APINODE.Count:%s", countStr)
	countInt, err := strconv.Atoi(countStr)
	if err != nil {
		log.Fatalf("convert countStr fail err:%v", err)
	}
	var addrs []string
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("APINODE.InternalAddr%d", i)
		addr := HttpPre + viper.GetString(addrName)
		addrs = append(addrs, addr)
	}

	// 启动多个节点
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("APINODE.InternalAddr%d", i)
		addr := HttpPre + viper.GetString(addrName)
		node := initNode(addr, addrs, leadFunc)
		go node.run()
	}
}
