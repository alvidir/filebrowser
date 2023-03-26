package user

import (
	"context"
	"encoding/json"

	fb "github.com/alvidir/filebrowser"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type UserEventPayload struct {
	UserID int32  `json:"id"`
	Kind   string `json:"kind"`
	UserProfile
}

type UserEventHandler struct {
	dirApp  *dir.DirectoryApplication
	fileApp *file.FileApplication
	logger  *zap.Logger
}

func NewUserEventHandler(dirApp *dir.DirectoryApplication, fileApp *file.FileApplication, logger *zap.Logger) *UserEventHandler {
	return &UserEventHandler{
		dirApp:  dirApp,
		fileApp: fileApp,
		logger:  logger,
	}
}

func (handler *UserEventHandler) OnEvent(ctx context.Context, body []byte) {
	event := new(UserEventPayload)
	if err := json.Unmarshal(body, event); err != nil {
		handler.logger.Error("unmarshaling user event body",
			zap.ByteString("event_body", body),
			zap.Error(err))

		return
	}

	switch kind := event.Kind; kind {
	case fb.EventKindCreated:
		handler.logger.Info("handling user event",
			zap.String("kind", kind))

		handler.onUserCreatedEvent(ctx, event)

	default:
		handler.logger.Warn("unhandled user event",
			zap.String("kind", kind))
	}
}

func (handler *UserEventHandler) onUserCreatedEvent(ctx context.Context, event *UserEventPayload) {
	if _, err := handler.dirApp.Create(ctx, event.UserID); err != nil {
		handler.logger.Error("creating directory",
			zap.Int32("user_id", event.UserID),
			zap.Error(err))

		return
	}

	data, err := json.Marshal(event.UserProfile)
	if err != nil {
		handler.logger.Error("marshaling user profile",
			zap.Error(err))

		return
	}

	_, err = handler.fileApp.Create(ctx, event.UserID, profilePath, data, nil)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("file_path", profilePath),
			zap.Int32("user_id", event.UserID),
			zap.ByteString("data", data),
			zap.Error(err))

		return
	}
}
