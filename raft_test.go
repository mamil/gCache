package main

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	log "github.com/sirupsen/logrus"
)

func TestClient(t *testing.T) {
	id := 0
	addrs := map[int]string{
		0: "http://localhost:8081",
	}
	raft := initNode(id, addrs, nil)
	receiveTimer := time.NewTimer(time.Duration((raft.heartBeat - 1000)) * time.Millisecond)
	raft.receiveTimer = receiveTimer

	message := make(chan string, MaxMessgaeSize)
	go raft.receive(addrs[id], message)
	time.Sleep(time.Duration(500) * time.Millisecond)
	raft.send("test from send", addrs[id])

	msg := <-message
	ast := assert.New(t)
	ast.Equal("test from send", msg)
}

func TestServerTimeout(t *testing.T) {
	id := 0
	addrs := map[int]string{
		0: "http://localhost:8081",
	}
	raft := initNode(id, addrs, nil)
	receiveTimer := time.NewTimer(time.Duration((raft.heartBeat - 1000)) * time.Millisecond)
	raft.receiveTimer = receiveTimer

	message := make(chan string, MaxMessgaeSize)
	go raft.receiveTimeoutHandler(message) // 开启接受超时

	msg := <-message
	log.Infof("message:%v", msg)
	ast := assert.New(t)
	ast.Equal("", msg)
}

func TestLeadElection(t *testing.T) {
	addrs := map[int]string{
		0: "http://localhost:8081",
		1: "http://localhost:8082",
		2: "http://localhost:8083",
	}

	called := uint64(0)
	leaderCall := func() {
		log.Infof("I'm leader!")
		atomic.AddUint64(&called, 1)
	}

	raft0 := initNode(0, addrs, leaderCall)
	go raft0.run()
	raft1 := initNode(1, addrs, leaderCall)
	go raft1.run()
	raft2 := initNode(2, addrs, leaderCall)
	go raft2.run()

	time.Sleep(time.Duration(20000) * time.Millisecond)
	log.Infof("### called:%d", called)
	ast := assert.New(t)
	ast.Equal(uint64(1), called)
}

// 单独组件测试
func TestTimer(t *testing.T) {
	heartbeat := 3000
	log.Infof("start timer")
	receiveTimer := time.NewTimer(time.Duration((heartbeat - 1000)) * time.Millisecond)

	go func(receiveTimer *time.Timer) {
		<-receiveTimer.C
		log.Infof("timeout!!")
	}(receiveTimer)

	time.Sleep(time.Duration(1) * time.Second)
	log.Infof("first sleep end, reset")
	receiveTimer.Reset(time.Duration((heartbeat - 1000)) * time.Millisecond)
	time.Sleep(time.Duration(1) * time.Second)
	log.Infof("2ed sleep end, reset")
	receiveTimer.Reset(time.Duration((heartbeat - 1000)) * time.Millisecond)

	time.Sleep(time.Duration(4) * time.Second)
}

func TestRandom(t *testing.T) {
	max := 700
	min := 300
	rand.Seed(time.Now().Unix())
	res := rand.Intn(max-min) + min
	log.Infof("res:%d", res)
}
