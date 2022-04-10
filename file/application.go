package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
}

type FileApplication struct {
	FileRepo FileRepository
	logger   *zap.Logger
}

func NewFileApplication(repo FileRepository, logger *zap.Logger) *FileApplication {
	return &FileApplication{
		FileRepo: repo,
		logger:   logger,
	}
}

func (app *FileApplication) Create(ctx context.Context, path string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"create\" file request",
		zap.Any(fb.AuthKey, ctx.Value(fb.AuthKey)),
		zap.String("path", path))

	return nil, nil
}
