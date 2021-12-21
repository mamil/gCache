package main

import (
	"net/http"

	"github.com/imroc/req"
	log "github.com/sirupsen/logrus"
)

const (
	// 节点角色
	Leader    = 0
	Follower  = 1
	Candidate = 2

	//
	MaxMessgaeSize = 50
)

type Raft struct {
	role int    // 现在的角色
	term uint64 // 任期

	addr     string   // 自己的地址
	peerAddr []string // 其他节点地址

	leadFunc LeaderFunc // 当选主节点之后的调用
}

func initNode(addr string, addrs []string, leadFunc LeaderFunc) *Raft {
	raft := &Raft{
		role:     Follower,
		term:     0,
		addr:     addr,
		peerAddr: addrs,
		leadFunc: leadFunc,
	}
	log.Infof("initNode:%+v", raft)
	return raft
}

func send(data string, addr string) {
	req.Post(addr, data)
	log.Infof("send[%+v] to %s", data, addr)
}

func receive(addr string, messageChan chan string) {
	http.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			len := r.ContentLength
			body := make([]byte, len)
			r.Body.Read(body)
			messageChan <- string(body)
			log.Infof("receive body[%s]", string(body))
		}))
	log.Println("raft receive is running at", addr)
	log.Info(http.ListenAndServe(addr[7:], nil))
}

func (r *Raft) run() {
	message := make(chan string, MaxMessgaeSize)
	go receive(r.addr, message)
	for {
		data := <-message

		if r.role == Follower {
			r.followerHandle(data)
		} else if r.role == Leader {
			r.leaderHandle(data)
		} else if r.role == Candidate {
			r.candidateHandle(data)
		}
	}
}

func (r *Raft) followerHandle(data string) {

}

func (r *Raft) leaderHandle(data string) {

}

func (r *Raft) candidateHandle(data string) {

}

// func heartbeat() {

// }
