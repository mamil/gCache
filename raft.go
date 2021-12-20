package main

const (
	// 节点角色
	Leader    = 0
	Follower  = 1
	Candidate = 2
)

type Raft struct {
	role int    // 现在的角色
	term uint64 // 任期
}

func initNode(){
	raft := &Raft{
		role: Follower,
	}
}

func heartbeat() {
	if 
}
