package freeLockCache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type MyLoader struct {
	cnt int32
}

func (m *MyLoader) Load(_ context.Context, _ []string) (map[string][]byte, error) {
	atomic.AddInt32(&m.cnt, 1)
	rtn := make(map[string][]byte, 3)
	rtn["1"] = []byte("a")
	rtn["2"] = []byte("b")
	rtn["3"] = []byte("c")
	time.Sleep(5)
	return rtn, nil
}

func TestCache_Get(t *testing.T) {
	loader := MyLoader{}
	cache, err := New(&Config{
		Enable:     true,
		DataLoader: &loader,
		Config:     bigcache.DefaultConfig(10 * time.Minute),
	})
	assert.Nil(t, err)

	ctx := context.Background()
	apps := []string{"1", "2", "3"}
	group := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		group.Add(1)
		go func() {
			group.Done()
			get, err := cache.Get(ctx, apps)
			assert.Nil(t, err)
			assert.Equal(t, []byte("a"), get["1"])
			assert.Equal(t, []byte("b"), get["2"])
			assert.Equal(t, []byte("c"), get["3"])
		}()
	}
	group.Wait()
	assert.Equal(t, int32(1), loader.cnt)
}

func TestCache_GetNoCache(t *testing.T) {
	loader := MyLoader{}
	cache, err := New(&Config{
		Enable:     false,
		DataLoader: &loader,
		Config:     bigcache.DefaultConfig(10 * time.Minute),
	})
	assert.Nil(t, err)

	ctx := context.Background()
	apps := []string{"1", "2", "3"}
	group := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		group.Add(1)
		go func() {
			group.Done()
			get, err := cache.Get(ctx, apps)
			assert.Nil(t, err)
			assert.Equal(t, []byte("a"), get["1"])
			assert.Equal(t, []byte("b"), get["2"])
			assert.Equal(t, []byte("c"), get["3"])
		}()
	}
	group.Wait()
	assert.Equal(t, int32(100), loader.cnt)
}
