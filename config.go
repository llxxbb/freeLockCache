package freeLockCache

import (
	"context"
	"github.com/allegro/bigcache/v3"
)

type Config struct {
	Enable bool
	DataLoader
	bigcache.Config
}
type DataLoader interface {
	Load(ctx context.Context, keys []string) (map[string][]byte, error)
}
