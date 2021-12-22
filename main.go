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
	"log"
	"strconv"

	"github.com/spf13/viper"
)

const (
	HttpPre = "http://"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
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

	apiAddr := HttpPre + viper.GetString("APINODE.InterfaceAddr")

	countStr := viper.GetString("CACHENODE.Count")
	log.Printf("CACHENODE.Count:%s", countStr)
	countInt, err := strconv.Atoi(countStr)
	if err != nil {
		log.Fatalf("convert countStr fail err:%v", err)
	}

	var addrs []string
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("CACHENODE.Addr%d", i)
		addr := HttpPre + viper.GetString(addrName)
		addrs = append(addrs, addr)
	}

	gcache := createGroup()
	if api {
		go startApiNode(apiAddr, gcache)
	}

	startAddr := HttpPre + viper.GetString(fmt.Sprintf("CACHENODE.Addr%d", node))
	startCacheServer(startAddr, addrs, gcache)
}
