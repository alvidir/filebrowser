package file

import (
	"context"
	"encoding/json"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	"go.uber.org/zap"
)

type FileEventHandler struct {
	fileApp *FileApplication
	certApp *cert.CertificateApplication
	issuers map[string]bool
	logger  *zap.Logger
}

func NewFileEventHandler(fileApp *FileApplication, certApp *cert.CertificateApplication, logger *zap.Logger) *FileEventHandler {
	return &FileEventHandler{
		fileApp: fileApp,
		certApp: certApp,
		issuers: make(map[string]bool),
		logger:  logger,
	}
}

func (handler *FileEventHandler) DiscardIssuer(issuer string) {
	handler.issuers[issuer] = false
}

func (handler *FileEventHandler) isDiscarted(issuer string) bool {
	accepted, exists := handler.issuers[issuer]
	return !accepted && exists
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
	case fb.EventKindCreated:
		handler.logger.Info("handling file event",
			zap.String("kind", kind))

		handler.onFileCreatedEvent(ctx, event)

	default:
		handler.logger.Warn("unhandled file event",
			zap.String("kind", kind))
	}
}

func (handler *FileEventHandler) onFileCreatedEvent(ctx context.Context, event *FileEventPayload) {
	if handler.isDiscarted(event.Issuer) {
		handler.logger.Info("discarting event",
			zap.String("issuer", event.Issuer),
			zap.String("app", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID))

		return
	}

	meta := Metadata{
		MetadataAppKey: event.AppID,
	}

	file, err := handler.fileApp.Create(ctx, event.UserID, event.FileName, nil, meta)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("issuer", event.Issuer),
			zap.String("app", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}

	_, err = handler.certApp.CreateFileAccessCertificate(ctx, event.UserID, file)
	if err != nil {
		handler.logger.Error("creating file access certificate",
			zap.String("issuer", event.Issuer),
			zap.String("app", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}
