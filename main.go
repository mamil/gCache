package main

/*
$ curl "http://localhost:9999/api?key=Tom"
630
$ curl "http://localhost:9999/api?key=kkk"
kkk not exist
*/

import (
	"flag"
	"fmt"
	"gCache/gcache"
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/viper"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

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

func initConfig() error {
	viper.SetConfigFile("./config.ini")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("read config fail, err:%v", err)
		return err
	}
	return nil
}

func main() {
	if err := initConfig(); err != nil {
		return
	}

	var node int
	var api bool
	flag.IntVar(&node, "node", 1, "gcache server node")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://" + viper.GetString("APINODE.Addr")

	countStr := viper.GetString("CACHENODE.Count")
	log.Printf("CACHENODE.Count:%s", countStr)
	countInt, err := strconv.Atoi(countStr)
	if err != nil {
		log.Fatalf("convert countStr fail err:%v", err)
	}

	var addrs []string
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("CACHENODE.Addr%d", i)
		addr := "http://" + viper.GetString(addrName)
		addrs = append(addrs, addr)
	}

	gcache := createGroup()
	if api {
		go startAPIServer(apiAddr, gcache)
	}

	startAddr := "http://" + viper.GetString(fmt.Sprintf("CACHENODE.Addr%d", node))
	startCacheServer(startAddr, addrs, gcache)
}
