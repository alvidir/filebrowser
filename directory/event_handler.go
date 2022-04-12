package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type DirectoryEventHandler struct {
	app    *DirectoryApplication
	logger *zap.Logger
}

func NewDirectoryEventHandler(app *DirectoryApplication, logger *zap.Logger) *DirectoryEventHandler {
	return &DirectoryEventHandler{
		app:    app,
		logger: logger,
	}
}

func (handler *DirectoryEventHandler) OnFileCreated(ctx context.Context, fileId, path string) {
	handler.logger.Info("processing a \"file created\" event",
		zap.Any(fb.AuthKey, ctx.Value(fb.AuthKey)))

	handler.app.AddFile(ctx, fileId, path, false)
}
