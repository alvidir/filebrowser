package file

import (
	"context"
	"encoding/json"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type FileEventHandler struct {
	fileApp *FileApplication
	issuers map[string]bool
	logger  *zap.Logger
}

func NewFileEventHandler(fileApp *FileApplication, logger *zap.Logger) *FileEventHandler {
	return &FileEventHandler{
		fileApp: fileApp,
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

	if handler.isDiscarted(event.Issuer) {
		handler.logger.Info("discarting event",
			zap.String("issuer", event.Issuer))

		return
	}

	switch kind := event.Kind; kind {
	case fb.EventKindCreated:
		handler.onFileCreatedEvent(ctx, event)

	case fb.EventKindDeleted:
		handler.onFileDeletedEvent(ctx, event)

	default:
		handler.logger.Warn("unhandled file event",
			zap.String("kind", event.Kind))
	}
}

func (handler *FileEventHandler) onFileCreatedEvent(ctx context.Context, event *FileEventPayload) {
	if len(event.Reference) > 0 {
		// if reference is set the file already exists
		handler.onFileUpdatedEvent(ctx, event)
		return
	}

	handler.logger.Info("handling a file \"created\" event")

	options := CreateOptions{
		Name: event.FileName,
		Meta: Metadata{
			MetadataAppKey: event.AppID,
			MetadataRefKey: event.FileID,
		},
	}

	_, err := handler.fileApp.Create(ctx, event.UserID, &options)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("issuer", event.Issuer),
			zap.String("app_id", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}

func (handler *FileEventHandler) onFileUpdatedEvent(ctx context.Context, event *FileEventPayload) {
	handler.logger.Info("handling a file \"updated\" event")

	options := UpdateOptions{
		Name: event.FileName,
		Meta: Metadata{
			MetadataAppKey: event.AppID,
			MetadataRefKey: event.FileID,
		},
	}

	_, err := handler.fileApp.Update(ctx, event.UserID, event.Reference, &options)
	if err != nil {
		handler.logger.Error("updating file",
			zap.String("issuer", event.Issuer),
			zap.String("reference", event.Reference),
			zap.String("app_id", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}

func (handler *FileEventHandler) onFileDeletedEvent(ctx context.Context, event *FileEventPayload) {
	handler.logger.Info("handling a file \"deleted\" event")

	_, err := handler.fileApp.Delete(ctx, event.UserID, event.FileID)
	if err != nil {
		handler.logger.Error("deleting file",
			zap.String("issuer", event.Issuer),
			zap.String("app_id", event.AppID),
			zap.String("file_name", event.FileName),
			zap.String("file_id", event.FileID),
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}
}
