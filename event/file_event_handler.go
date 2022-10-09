package event

import (
	"context"
	"encoding/json"

	cert "github.com/alvidir/filebrowser/certificate"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type FileEventPayload struct {
	UserID   int32  `json:"user_id"`
	App      string `json:"app"`
	Url      string `json:"url"`
	FileName string `json:"name"`
	Kind     string `json:"kind"`
}

type FileEventHandler struct {
	dirApp  *dir.DirectoryApplication
	fileApp *file.FileApplication
	certApp *cert.CertificateApplication
	logger  *zap.Logger
}

func NewFileEventHandler(dirApp *dir.DirectoryApplication, fileApp *file.FileApplication, certApp *cert.CertificateApplication, logger *zap.Logger) *FileEventHandler {
	return &FileEventHandler{
		dirApp:  dirApp,
		fileApp: fileApp,
		certApp: certApp,
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
		file.MetadataAppKey: event.App,
		file.MetadataUrlKey: event.Url,
	}

	file, err := handler.fileApp.Create(ctx, event.UserID, event.FileName, nil, meta)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("url", event.Url),
			zap.String("app", event.App),
			zap.String("file_name", event.FileName),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}

	_, err = handler.certApp.CreateFileAccessCertificate(ctx, event.UserID, file)
	if err != nil {
		handler.logger.Error("creating file access certificate",
			zap.String("url", event.Url),
			zap.String("app", event.App),
			zap.String("file_name", event.FileName),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}
