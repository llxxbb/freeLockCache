package freeLockCache

import "context"

type Config struct {
	Enable bool
	DataLoader
}
type DataLoader interface {
	Load(ctx context.Context, keys []string) (map[string][]byte, error)
}
