package freeLockCache

import (
	"context"
	"errors"
	"github.com/allegro/bigcache/v3"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"sync"
)

type Cache struct {
	*Config
	locks   map[string]*sync.Mutex
	keyLock sync.Mutex
	bc      *bigcache.BigCache
}

// Get GetDN 获取App对应的域名
func (c *Cache) Get(ctx context.Context, keys []string) (map[string][]byte, error) {
	// check
	keyNum := len(keys)
	if keyNum == 0 {
		return nil, nil
	}

	if !c.Enable {
		// load data directly
		return c.Load(ctx, keys)
	}
	// get from cache
	rtn, noCache := c.getFromCache(keys)
	noCacheLen := len(noCache)
	if noCacheLen == 0 {
		return rtn, nil
	}

	// load data
	err := c.loadWithLock(ctx, keys)
	if err != nil {
		return nil, err
	}

	// get rest
	rtn2, noCache := c.getFromCache(noCache)
	// append rtn2 to rtn
	for k, v := range rtn2 {
		rtn[k] = v
	}

	// check, because cache loading may cause error
	if len(rtn) < keyNum {
		msg := "cache item missed"
		zap.L().Error(msg)
		return nil, errors.New(msg)
	}

	return rtn, nil
}

func (c *Cache) loadWithLock(ctx context.Context, keys []string) error {
	key := keys[0]
	c.keyLock.Lock()
	mLock, ok := c.locks[key]
	if ok {
		mLock.Lock()
		mLock.Unlock()
		// 等完后不需要做任何事
		return nil
	}
	// 创建锁，防止并发处理 keyLock.Unlock() 后面的逻辑
	oneLock := sync.Mutex{}
	oneLock.Lock()
	defer oneLock.Unlock()
	c.locks[key] = &oneLock

	c.keyLock.Unlock()

	// 加载数据--------------------------------------
	// 加载完成后删除锁
	defer delete(c.locks, key)

	err := c.loadToCache(ctx, keys)
	return err
}

func (c *Cache) loadToCache(ctx context.Context, keys []string) error {
	// get appInfo
	dn, err := c.Load(ctx, keys)
	if err != nil {
		return err
	}

	// set to catch and result
	for k, v := range dn {
		err := c.bc.Set(k, v)
		if err != nil {
			zap.Error(err)
			return err
		}
	}
	return nil
}

// 从缓存中获取域名信息，返回已缓存的和未
func (c *Cache) getFromCache(keys []string) (map[string][]byte, []string) {
	rtn := make(map[string][]byte)
	noCache := make(map[string]struct{})

	// get from cache
	for _, k := range keys {
		s, err := c.bc.Get(k)
		if err == nil {
			rtn[k] = s
		} else {
			noCache[k] = struct{}{}
		}
	}
	return rtn, maps.Keys(noCache)
}

func New(cfg *Config) (*Cache, error) {
	c := Cache{
		Config:  cfg,
		locks:   make(map[string]*sync.Mutex),
		keyLock: sync.Mutex{},
	}
	var err error
	c.bc, err = bigcache.New(context.Background(), cfg.Config)
	return &c, err
}
