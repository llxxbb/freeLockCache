package freeLockCache

import (
	"context"
	"github.com/allegro/bigcache/v3"
)

type Config[K comparable, V any] struct {
	Enable bool
	Load   func(ctx context.Context, key K) (V, error)
	bigcache.Config
}
