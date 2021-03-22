package singleflight

import "sync"

// 正在进行中的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// 记录有哪些key正在请求
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// 同时来多个相同key的请求的时候，合并处理，只调用一次
// 但不会把长时间的请求都合并
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call) // 延迟创建
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新 g.m。 不能一直留着，数据可能过期，更新，下次要重新获取
	g.mu.Unlock()

	return c.val, c.err
}
