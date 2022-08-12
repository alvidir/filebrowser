package permissions

import (
	"context"

	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type PermissionsRepository interface {
	Find(ctx context.Context, fileId string) (*file.Permissions, error)
}

type PermissionsApplication struct {
	permissionsRepo PermissionsRepository
	logger          *zap.Logger
}

func NewPermissionsApplication(permissionsRepo PermissionsRepository, fileRepo file.FileRepository, logger *zap.Logger) *PermissionsApplication {
	return &PermissionsApplication{
		permissionsRepo: permissionsRepo,
		logger:          logger,
	}
}

func (app *PermissionsApplication) HasPermissions(ctx context.Context, uid int32, fileId string) error {
	app.logger.Info("processing a \"create\" Permissions request",
		zap.Int32("user_id", uid))

	return nil
}
