package freeLockCache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var cnt int32

func Load(_ context.Context, key string) (string, error) {
	cnt++
	time.Sleep(5)
	if key == "hello" {
		return "world", nil
	}
	return "", nil
}

func TestCache_Get(t *testing.T) {
	cnt = 0
	cache, err := New(&Config[string, string]{
		Enable: true,
		Load:   Load,
		Config: bigcache.DefaultConfig(10 * time.Minute),
	})
	assert.Nil(t, err)

	ctx := context.Background()
	group := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		group.Add(1)
		go func() {
			group.Done()
			get, err := cache.Get(ctx, "hello")
			assert.Nil(t, err)
			assert.Equal(t, "world", get)
		}()
	}
	group.Wait()
	// 只加载1次
	assert.Equal(t, int32(1), cnt)
}

func TestCache_GetNoCache(t *testing.T) {
	cnt = 0
	cache, err := New(&Config[string, string]{
		Enable: false,
		Load:   Load,
		Config: bigcache.DefaultConfig(10 * time.Minute),
	})
	assert.Nil(t, err)

	ctx := context.Background()
	group := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		group.Add(1)
		go func() {
			group.Done()
			get, err := cache.Get(ctx, "lxb")
			assert.Nil(t, err)
			assert.Equal(t, "", get)
		}()
	}
	group.Wait()
	assert.Equal(t, int32(100), cnt)
}
