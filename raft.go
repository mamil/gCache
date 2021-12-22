package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

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

// 内部节点之间的通信
type InternalMessage struct {
	MsgType       string // 消息类型
	Addr          string // 消息发送者地址, 这里不存发送者id，这样就不需要配置完全一样
	Term          uint64 // 任期
	VoteGranted   bool   // 是否投票
	CandidateAddr string // 候选人地址
	LastLogIndex  uint64 // 最新的log id
	LastLogTerm   uint64 // 最新的任期
}
type Raft struct {
	role      int     // 现在的角色
	term      uint64  // 任期
	logIndex  uint64  // log id, 这里仅用到选举流程，不用这个也没关系
	heartBeat int64   // 心跳时间,ms
	waitMs    []int64 // 随机等待时间,ms
	timerGap  int64   // timer设置的时间

	nextLeaderElectionTime int64  // 下次选举时间,ms,13位时间戳
	gettedVote             int    // 获取的选票数
	voteFor                string // 投票给了谁
	voteForTerm            uint64 // 上次投票的任期

	receiveTimer *time.Timer //接收超时计时器

	id       int            // 自己的地址
	peerAddr map[int]string // 其他节点地址

	leadFunc LeaderFunc // 当选主节点之后的调用
}

func initNode(id int, addrs map[int]string, leadFunc LeaderFunc) *Raft {
	raft := &Raft{
		role:      Follower,
		term:      0,
		heartBeat: 3000,
		waitMs:    []int64{300, 700},
		id:        id,
		peerAddr:  addrs,
		leadFunc:  leadFunc,
	}
	raft.timerGap = raft.heartBeat / 2
	log.Infof("initNode:%+v", raft)
	return raft
}

func (r *Raft) send(data string, addr string) {
	req.Post(addr, data)
	log.Infof("send self:%s to %s, data[%+v]", r.peerAddr[r.id], addr, data)
}

// 接收peer node的内部消息
func (r *Raft) receive(addr string, messageChan chan string) {
	log.Infof("receive addr:%s", addr)
	handler := func(w http.ResponseWriter, req *http.Request) {
		r.resetReceiveTimer()

		len := req.ContentLength
		body := make([]byte, len)
		req.Body.Read(body)
		messageChan <- string(body)
		log.Infof("receive self:%s, body[%s]", r.peerAddr[r.id], string(body))
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.HandlerFunc(handler))
	log.Println("raft receive is running at", addr)
	log.Info(http.ListenAndServe(addr[7:], serveMux))
}

// 接收超时处理
func (r *Raft) receiveTimeoutHandler(messageChan chan string) {
	for {
		<-r.receiveTimer.C

		// 超时，向chan发送空字符串
		emptyStr := ""
		messageChan <- emptyStr
		r.resetReceiveTimer()
	}
}

// 重设timer，单位ms
func (r *Raft) resetReceiveTimer() {
	r.receiveTimer.Reset(time.Duration(r.timerGap+int64(rand.Int63n(r.timerGap))) * time.Millisecond)
}

// 计算下一次选举时间
func (r *Raft) getNextLeaderElectionTime() int64 {
	rand.Seed(time.Now().Unix())
	return time.Now().UnixNano()/1e6 + int64(r.heartBeat) + int64(rand.Int63n(r.waitMs[1]-r.waitMs[0])) + r.waitMs[0]
}

func (r *Raft) run() {
	message := make(chan string, MaxMessgaeSize)
	receiveTimer := time.NewTimer(time.Duration(r.timerGap+int64(rand.Int63n(r.timerGap))) * time.Millisecond)
	r.receiveTimer = receiveTimer
	r.nextLeaderElectionTime = r.getNextLeaderElectionTime()
	go r.receiveTimeoutHandler(message)     // 开启接受超时
	go r.receive(r.peerAddr[r.id], message) // 开启接收服务
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
	log.Infof("followerHandle addr:%s data:%s", r.peerAddr[r.id], data)

	if data == "" {
		currentTime := time.Now().UnixNano() / 1e6
		log.Infof("followerHandle addr:%s, currentTime:%d, nextLeaderElectionTime:%d", r.peerAddr[r.id], currentTime, r.nextLeaderElectionTime)
		if currentTime > r.nextLeaderElectionTime { // 开始选举
			r.role = Candidate
			r.term += 1
			r.voteFor = r.peerAddr[r.id]
			r.gettedVote = 1
			r.voteForTerm = r.term
			r.requestVote()
			log.Infof("followerHandle change to Candidate, request vote, addr:%s", r.peerAddr[r.id])
		}
	} else {
		var message InternalMessage
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			log.Errorf("followerHandle Unmarshal fail, data:%v", data)
			return
		}
		if message.MsgType == "RequestVote" {
			r.responseVote(message)
			r.role = Follower
		} else if message.MsgType == "AppendEntries" {
			r.role = Follower
			r.nextLeaderElectionTime = r.getNextLeaderElectionTime()
		}
	}
}

func (r *Raft) leaderHandle(data string) {
	log.Infof("leaderHandle I'm leader, addr:%s", r.peerAddr[r.id])

	message := InternalMessage{
		MsgType: "AppendEntries",
		Addr:    r.peerAddr[r.id],
	}

	dataByte, err := json.Marshal(&message)
	if err != nil {
		log.Errorf("leaderHandle Marshal message fail, err:%v, message:%+v", err, message)
		return
	}
	dataStr := string(dataByte[:])

	for index, peerAddr := range r.peerAddr {
		if index == r.id {
			continue
		}
		r.send(dataStr, peerAddr)
	}
}

func (r *Raft) candidateHandle(data string) {
	log.Infof("candidateHandle I'm candidate, addr:%s", r.peerAddr[r.id])

	if data != "" {
		var message InternalMessage
		if err := json.Unmarshal([]byte(data), &message); err != nil {
			log.Errorf("followerHandle %s Unmarshal fail, data:%v", r.peerAddr[r.id], data)
			return
		}

		if message.MsgType == "responseVote" && message.VoteGranted {
			r.gettedVote += 1
			log.Infof("candidateHandle %s now have vote:%d, need:%d", r.peerAddr[r.id], r.gettedVote, len(r.peerAddr)/2)

			if r.gettedVote > len(r.peerAddr)/2 {
				r.gettedVote = 0
				r.role = Leader
				r.leadFunc()
				log.Infof("candidateHandle %s change to leader", r.peerAddr[r.id])
			}
		} else if message.MsgType == "RequestVote" {
			r.responseVote(message)
		} else if message.MsgType == "AppendEntries" {
			r.gettedVote = 0
			r.role = Follower
			r.nextLeaderElectionTime = r.getNextLeaderElectionTime() // 重置下次选举时间
			log.Infof("candidateHandle %s change to follower", r.peerAddr[r.id])
		}
	} else { // 一次选举超时，开始下一次
		log.Infof("candidateHandle %s timeout, term mod to:%d", r.peerAddr[r.id], r.term+1)

		r.term += 1
		r.voteFor = r.peerAddr[r.id]
		r.gettedVote = 1
		r.voteForTerm = r.term
		r.requestVote()
	}
}

func (r *Raft) requestVote() {
	r.voteFor = r.peerAddr[r.id] // 投票给自己
	message := InternalMessage{
		MsgType:       "RequestVote",
		Addr:          r.peerAddr[r.id],
		Term:          r.term,
		CandidateAddr: r.peerAddr[r.id],
		LastLogIndex:  r.logIndex,
		LastLogTerm:   r.term,
	}

	dataByte, err := json.Marshal(&message)
	if err != nil {
		log.Errorf("requestVote Marshal message fail, err:%v, message:%+v", err, message)
		return
	}
	dataStr := string(dataByte[:])

	for index, peerAddr := range r.peerAddr {
		if index == r.id {
			continue
		}
		r.send(dataStr, peerAddr)
	}
}

func (r *Raft) responseVote(message InternalMessage) {
	response := InternalMessage{
		MsgType:     "responseVote",
		Addr:        r.peerAddr[r.id],
		Term:        r.term,
		VoteGranted: false,
	}

	if message.Term < r.term {
		response.VoteGranted = false
	} else if (r.voteFor == "" || r.voteFor == message.Addr) && message.LastLogIndex >= r.logIndex {
		r.voteFor = message.Addr
		r.voteForTerm = message.LastLogTerm

		response.VoteGranted = true
		log.Infof("responseVote self:%s, vote to peer:%s, r.logIndex:%d, message.LastLogIndex:%d",
			r.peerAddr[r.id], message.Addr, r.logIndex, message.LastLogIndex)

	} else if r.voteFor != "" && message.Term > r.voteForTerm { // 上次选举失败，有更新的开始，则投票
		r.voteFor = message.Addr
		r.voteForTerm = message.Term

		response.VoteGranted = true
		log.Infof("responseVote self:%s, vote to peer:%s, r.voteForTerm:%d, message.Term:%d",
			r.peerAddr[r.id], message.Addr, r.voteForTerm, message.Term)
	}

	// 回复
	dataByte, err := json.Marshal(&response)
	if err != nil {
		log.Errorf("responseVote Marshal message fail, err:%v, message:%+v", err, response)
		return
	}
	dataStr := string(dataByte[:])
	r.send(dataStr, message.Addr)
}
