package permissions

import (
	"context"

	"github.com/alvidir/filebrowser/file"
	"github.com/go-redis/redis/v9"
	"go.uber.org/zap"
)

type RedisPermissionsRepository struct {
	conn     *redis.Client
	fileRepo file.FileRepository
	logger   *zap.Logger
}

// NewRedisPermissionsRepository returns an implementation of Cache for RedisPermissionsRepository
func NewRedisPermissionsRepository(conn *redis.Client, fileRepo file.FileRepository, logger *zap.Logger) *RedisPermissionsRepository {
	return &RedisPermissionsRepository{
		conn:     conn,
		fileRepo: fileRepo,
	}
}

func (c *RedisPermissionsRepository) Find(ctx context.Context, fileId string) (*file.Permissions, error) {
	return nil, nil
}
