package main

import (
	"math/rand"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// func TestClient(t *testing.T) {
// 	message := make(chan string, MaxMessgaeSize)
// 	addr := "http://localhost:8081"
// 	go receive(addr, message)
// 	time.Sleep(time.Duration(5) * time.Second)
// 	send("test from send", addr)
// }

// func TestServerTimeout(t *testing.T) {
// 	message := make(chan string, MaxMessgaeSize)
// 	addr := "http://localhost:8081"
// 	go receive(addr, message)
// 	time.Sleep(time.Duration(5) * time.Second)
// }

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
