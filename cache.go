package freeLockCache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/goccy/go-json"
	"sync"
)

type Cache[K comparable, V any] struct {
	*Config[K, V]
	locks   map[string]*sync.Mutex
	bigLock sync.Mutex
	bc      *bigcache.BigCache
}

// Get GetDN 获取App对应的域名
func (c *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	// check
	var rtn V
	if !c.Enable {
		// load data directly
		return c.Load(ctx, key)
	}
	// convert key
	keyB, err := json.Marshal(key)
	if err != nil {
		return rtn, err
	}
	keyS := string(keyB)
	if keyS == "" {
		return rtn, nil
	}

	// get from cache
	rtn, err = c.getFromCache(keyS)
	if err == nil {
		return rtn, nil
	}

	// load data and return
	return c.loadWithLock(ctx, keyS, key)
}

func (c *Cache[K, V]) loadWithLock(ctx context.Context, keyS string, key K) (V, error) {
	c.bigLock.Lock()
	// 如果存在 keyLock 则认为正在加载，直接等待结束并返回结果
	keyLock, ok := c.locks[keyS]
	if ok {
		c.bigLock.Unlock()
		// 等待从外部加载到缓存完
		keyLock.Lock()
		// 加载完成后释放键上的锁
		keyLock.Unlock()
		// 返回缓存数据
		return c.getFromCache(keyS)
	}
	// 创建锁，防止并发处理 bigLock.Unlock() 后面的逻辑
	oneLock := sync.Mutex{}
	oneLock.Lock()
	c.locks[keyS] = &oneLock
	defer oneLock.Unlock()
	c.bigLock.Unlock()

	// 加载数据--------------------------------------
	// 加载完成后删除锁
	defer delete(c.locks, keyS)

	return c.loadToCache(ctx, keyS, key)
}

func (c *Cache[K, V]) loadToCache(ctx context.Context, keyS string, key K) (V, error) {
	// get appInfo
	rtn, err := c.Load(ctx, key)
	if err != nil {
		return rtn, err
	}

	// set to catch and result
	v, err := serialize(rtn)
	err = c.bc.Set(keyS, v)
	if err != nil {
		return rtn, err
	}
	return rtn, nil
}

// 从缓存中获取域名信息，返回已缓存的和未, 如果未缓存则返回 error
func (c *Cache[K, V]) getFromCache(key string) (V, error) {
	var rtn V
	// get from cache, no cache then get error
	s, err := c.bc.Get(key)
	if err != nil {
		return rtn, err
	}
	err = deserialize(s, &rtn)
	return rtn, err
}

func New[K comparable, V any](cfg *Config[K, V]) (*Cache[K, V], error) {
	c := Cache[K, V]{
		Config:  cfg,
		locks:   make(map[string]*sync.Mutex),
		bigLock: sync.Mutex{},
	}
	var err error
	c.bc, err = bigcache.New(context.Background(), cfg.Config)
	return &c, err
}
