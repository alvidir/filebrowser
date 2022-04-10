package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type DirectoryRepository interface {
	Create(ctx context.Context, directory *Directory) error
}

type DirectoryApplication struct {
	directoryRepo DirectoryRepository
	logger        *zap.Logger
}

func NewDirectoryApplication(repo DirectoryRepository, logger *zap.Logger) *DirectoryApplication {
	return &DirectoryApplication{
		directoryRepo: repo,
		logger:        logger,
	}
}

func (app *DirectoryApplication) Create(ctx context.Context) (*Directory, error) {
	app.logger.Info("processing a \"create\" directory request",
		zap.Any(fb.AuthKey, ctx.Value(fb.AuthKey)))

	uid, ok := ctx.Value(fb.AuthKey).(int32)
	if !ok {
		app.logger.Error("asserting authentication id",
			zap.Any(fb.AuthKey, ctx.Value(fb.AuthKey)))
		return nil, fb.ErrUnauthorized
	}

	directory := NewDirectory(uid)
	if err := app.directoryRepo.Create(ctx, directory); err != nil {
		return nil, err
	}

	return directory, nil
}
