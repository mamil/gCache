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
	"gcache/cache"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	HttpPre       = "http://"
	DefaultNodeId = 0
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

func InitLogger() {
	level := GetLogLevel(viper.GetString("LOG_LEVEL"))
	log.SetLevel(level)
	formatter := &log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		// DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05.999999999",
	}
	log.SetFormatter(formatter)
	log.Debug("debug log level")
	log.Info("start")
}

func GetLogLevel(logLevelConfig string) log.Level {
	level := log.InfoLevel

	logLevelConfig = strings.ToUpper(logLevelConfig)

	switch logLevelConfig {
	case "DEBUG":
		level = log.DebugLevel
	case "INFO":
		level = log.InfoLevel
	case "ERROR":
		level = log.ErrorLevel
	case "FATAL":
		level = log.FatalLevel
	case "TRACE":
		level = log.TraceLevel
	case "WARN":
		level = log.WarnLevel
	}
	return level
}

func NodePrepare(nodeId int) (startAddr string, addrs []string, g *cache.Group) {
	// 获取节点地址
	countStr := viper.GetString("CACHENODE.Count")
	log.Infof("CACHENODE.Count:%s", countStr)
	countInt, err := strconv.Atoi(countStr)
	if err != nil {
		log.Fatalf("convert countStr fail err:%v", err)
	}
	// var addrs []string
	for i := 0; i < countInt; i++ {
		addrName := fmt.Sprintf("CACHENODE.Addr%d", i)
		addr := HttpPre + viper.GetString(addrName)
		addrs = append(addrs, addr)
	}

	g = createGroup()
	startAddr = HttpPre + viper.GetString(fmt.Sprintf("CACHENODE.Addr%d", nodeId))
	return
}

func StartNode(startAddr string, addrs []string, g *cache.Group) {
	startCacheServer(startAddr, addrs, g)
}

func main() {
	if err := initConfig(); err != nil {
		return
	}

	InitLogger()

	var node int
	var api int
	flag.IntVar(&node, "node", -1, "gcache server node")
	flag.IntVar(&api, "api", -1, "gcache api node")
	flag.Parse()

	// 此节点是api还是node
	if node >= 0 {
		startAddr, addrs, g := NodePrepare(node)
		StartNode(startAddr, addrs, g)
	}
	if api >= 0 {
		apiAddr := HttpPre + viper.GetString("APINODE.InterfaceAddr")
		log.Infof("apiAddr:%s", apiAddr)
		startApiNode(api, apiAddr, DefaultNodeId)
	}
}
