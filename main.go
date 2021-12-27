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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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

func main() {
	if err := initConfig(); err != nil {
		return
	}

	InitLogger()

	var node int
	var api bool
	flag.IntVar(&node, "node", 1, "gcache server node")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := HttpPre + viper.GetString("APINODE.InterfaceAddr")

	countStr := viper.GetString("CACHENODE.Count")
	log.Infof("CACHENODE.Count:%s", countStr)
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
