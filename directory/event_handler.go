package directory

import (
	"context"
	"time"

	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 10_000 * time.Millisecond
)

type DirectoryEventHandler struct {
	Timeout time.Duration
	app     *DirectoryApplication
	logger  *zap.Logger
}

func NewDirectoryEventHandler(app *DirectoryApplication, logger *zap.Logger) *DirectoryEventHandler {
	return &DirectoryEventHandler{
		app:    app,
		logger: logger,
	}
}

func (handler *DirectoryEventHandler) OnFileCreated(file *file.File, uid int32, path string) {
	handler.logger.Info("processing a \"file created\" event",
		zap.Any("user_id", uid))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler.app.AddFile(ctx, file, uid, path)
}
