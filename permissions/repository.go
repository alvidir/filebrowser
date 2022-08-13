package permissions

import (
	"context"

	"github.com/alvidir/filebrowser/file"
	"github.com/go-redis/cache/v8"
	"go.uber.org/zap"
)

type RedisPermissionsRepository struct {
	cache    *cache.Cache
	fileRepo file.FileRepository
	logger   *zap.Logger
}

// NewRedisPermissionsRepository returns an implementation of Cache for RedisPermissionsRepository
func NewRedisPermissionsRepository(cache *cache.Cache, fileRepo file.FileRepository, logger *zap.Logger) *RedisPermissionsRepository {
	return &RedisPermissionsRepository{
		cache:    cache,
		fileRepo: fileRepo,
	}
}

func (c *RedisPermissionsRepository) Find(ctx context.Context, fileId string) (*file.Permissions, error) {
	return nil, nil
}
