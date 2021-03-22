package gcache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string
	mainCache cache
	getter    Getter
	peers     PeerPicker
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 创建新的缓存组
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("Getter is nil")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		mainCache: cache{cacheBytes: cacheBytes},
		getter:    getter,
	}
	groups[name] = g
	log.Printf("new group %s added", name)
	return g
}

// 根据组名，获取缓存组
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 获取键值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}

	if v, ok := g.mainCache.get(key); ok { // 先查找缓存，是否有这个数据
		log.Println("gcache hit")
		return v, nil
	}

	return g.load(key) // 缓存没有，就从本地加载到cache
}

// 使用 PickPeer() 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。若是本机节点或失败，则回退到 getLocal()，从本地数据库加载数据。
func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}

	return g.getLocal(key)
}

// 以用户定义的方式从本地加载数据
func (g *Group) getLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 调用用户定义的方法
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneByte(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 把数据加到缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 注册其他缓存服务
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

// 从其他缓存服务获取数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
