package main

import (
	"fmt"
	"gcache/cache"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LeaderFunc func()

// 启动对外API服务
// 上一个API服务如果是没有释放端口的，需要一种清理机制
func startAPIServer(apiAddr string, g *cache.Group) LeaderFunc {
	return func() {
		//清理
		cleanPort(apiAddr)

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
		log.Fatal(http.ListenAndServe(apiAddr[7:], serveMux))
	}
}

func startApiNode(apiAddr string, g *cache.Group) {
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

func cleanPort(apiAddr string) {
	log.Infof("cleanPort apiAddr:%s", apiAddr)

	portStr := strings.Split(apiAddr, ":")[2]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Errorf("cleanPort Atoi port fail, portStr:%s, err:%v", portStr, err)
		return
	}
	log.Infof("cleanPort port:%d", port)

	checkStatement := fmt.Sprintf(`netstat -anp | grep -q %d ; echo $?`, port)
	output, err := exec.Command("sh", "-c", checkStatement).CombinedOutput()
	if err != nil {
		log.Errorf("cleanPort run Command fail, err:%v", err)
		return
	}
	log.Debugf("cleanPort run Command result:%s", string(output))

	// maybe return with :(Not all processes could be identified, non-owned process info\n will not be shown, you would have to be root to see it all.)\n1\n
	res := string(output)
	if strings.Contains(string(output), "Not all processes") {
		resTmp := strings.Split(string(output), "\n")
		res = resTmp[2]
	}

	// log.Println(output, string(output)) ==> [48 10] 0 或 [49 10] 1
	result, err := strconv.Atoi(strings.TrimSuffix(res, "\n"))
	if err != nil {
		log.Errorf("cleanPort Atoi output fail, portStr:%s, err:%v", portStr, err)
		return
	}
	if result == 0 {
		log.Infof("cleanPort port:%d in use", port)
	} else {
		log.Infof("cleanPort port:%d not in use", port)
	}
}
