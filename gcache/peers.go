package gcache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool) // 根据传入的 key 选择相应节点 PeerGetter
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error) // 从对应 group 查找缓存值
}
