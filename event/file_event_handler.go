package event

import (
	"context"
	"encoding/json"
	"fmt"

	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type FileEventPayload struct {
	UserID   int32  `json:"user_id"`
	FileID   string `json:"id"`
	FileName string `json:"name"`
	App      string `json:"app"`
	Kind     string `json:"kind"`
}

type FileEventHandler struct {
	dirApp  *dir.DirectoryApplication
	fileApp *file.FileApplication
	logger  *zap.Logger
}

func NewFileEventHandler(dirApp *dir.DirectoryApplication, fileApp *file.FileApplication, logger *zap.Logger) *FileEventHandler {
	return &FileEventHandler{
		dirApp:  dirApp,
		fileApp: fileApp,
		logger:  logger,
	}
}

func (handler *FileEventHandler) OnEvent(ctx context.Context, body []byte) {
	event := new(FileEventPayload)
	if err := json.Unmarshal(body, event); err != nil {
		handler.logger.Error("unmarshaling file event body",
			zap.ByteString("event_body", body),
			zap.Error(err))

		return
	}

	switch kind := event.Kind; kind {
	case EVENT_KIND_CREATED:
		handler.logger.Info("handling file event",
			zap.String("kind", kind))

		handler.onFileCreatedEvent(ctx, event)

	default:
		handler.logger.Warn("unhandled file event",
			zap.String("kind", kind))
	}
}

func (handler *FileEventHandler) onFileCreatedEvent(ctx context.Context, event *FileEventPayload) {
	meta := file.Metadata{
		file.MetadataOriginKey: fmt.Sprintf(
			file.FileOriginFormat,
			event.App,
			event.FileID,
		),
	}

	_, err := handler.fileApp.Create(ctx, event.UserID, event.FileName, nil, meta)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("file_id", event.FileID),
			zap.String("file_name", event.FileName),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}
