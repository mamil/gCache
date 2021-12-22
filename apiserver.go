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
// 上一个API服务如果是没有释放端口的，需要一种清理机制
func startAPIServer(apiAddr string, g *gcache.Group) LeaderFunc {
	return func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}

		serveMux := http.NewServeMux()
		serveMux.Handle("/api", http.HandlerFunc(handler))
		log.Fatal(http.ListenAndServe(apiAddr[7:], serveMux))
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
	addrs := map[int]string{}
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("APINODE.InternalAddr%d", i)
		addrs[i] = HttpPre + viper.GetString(addrName)
	}

	// 启动多个节点
	for i := 0; i < countInt; i++ {
		node := initNode(i, addrs, leadFunc)
		go node.run()
	}
}
