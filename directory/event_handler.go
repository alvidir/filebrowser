package directory

import (
	"context"

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

func (handler *DirectoryEventHandler) OnFileCreated(uid int32, fileId, path string) {
	handler.logger.Info("processing a \"file created\" event",
		zap.Any("uid", uid))

	ctx := context.Background()
	handler.app.AddFile(ctx, uid, fileId, path, false)
}
