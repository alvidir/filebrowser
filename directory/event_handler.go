package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type FileCreatedEvent interface {
	Context() context.Context
	FileId() string
	Path() string
}

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

func (handler *DirectoryEventHandler) onFileCreatedEvent(event FileCreatedEvent) {
	handler.logger.Info("processing a \"file created\" event",
		zap.Any(fb.AuthKey, event.Context().Value(fb.AuthKey)))

	handler.app.AddFile(event.Context(), event.FileId(), event.Path(), false)
}

func (handler *DirectoryEventHandler) handle(ctx context.Context, in chan interface{}) {
	handler.logger.Info("running directory's event handler")
	defer close(in)

	for {
		select {
		case event := <-in:
			if data, ok := event.(FileCreatedEvent); ok {
				handler.onFileCreatedEvent(data)
			} else {
				handler.logger.Error("handling event: unknown type")
			}
		case <-ctx.Done():
			handler.logger.Warn("terminating event handler",
				zap.Error(ctx.Err()))

			break
		}
	}
}

func (handler *DirectoryEventHandler) Run(ctx context.Context, size int) chan<- interface{} {
	inout := make(chan interface{}, size)
	go handler.handle(ctx, inout)
	return inout
}
