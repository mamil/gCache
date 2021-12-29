package main

import (
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LeaderFunc func()

// 启动对外API服务
func startAPIServer(apiAddr string, nodeId int) LeaderFunc {
	return func() {
		// 启动api下的node
		startAddr, addrs, g := NodePrepare(nodeId)
		go StartNode(startAddr, addrs, g)
		// log.Infof("##### g:%+v", *g)

		//监听
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
		log.Infof("startAPIServer on %s", apiAddr)
		log.Fatal(http.ListenAndServe(apiAddr[7:], serveMux))
	}
}

func startApiNode(apiId int, apiAddr string, nodeId int) {
	// 获取leader的回调函数
	leadFunc := startAPIServer(apiAddr, nodeId)

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

	// 启动节点
	node := initNode(apiId, addrs, leadFunc)
	node.run()
}
