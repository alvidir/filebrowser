package event

import (
	"context"
	"encoding/json"

	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

const (
	userProfilePath = ".profile"
)

type UserProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserEventPayload struct {
	ID   int32  `json:"id"`
	Kind string `json:"kind"`
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
		handler.logger.Error("unmarshaling event body",
			zap.ByteString("event_body", body),
			zap.Error(err))

		return
	}

	switch kind := event.Kind; kind {
	case EVENT_KIND_CREATED:
		handler.logger.Info("handling event",
			zap.String("kind", kind))

		handler.onUserCreatedEvent(ctx, event)

	default:
		handler.logger.Warn("unhandled event",
			zap.String("kind", kind))
	}
}

func (handler *UserEventHandler) onUserCreatedEvent(ctx context.Context, event *UserEventPayload) {
	if _, err := handler.dirApp.Create(ctx, event.ID); err != nil {
		handler.logger.Error("creating directory",
			zap.Int32("user_id", event.ID),
			zap.Error(err))

		return
	}

	data, err := json.Marshal(event.UserProfile)
	if err != nil {
		handler.logger.Error("marshaling event profile",
			zap.Error(err))

		return
	}

	_, err = handler.fileApp.Create(ctx, event.ID, userProfilePath, data, nil)
	if err != nil {
		handler.logger.Error("creating file",
			zap.String("file_path", userProfilePath),
			zap.Int32("user_id", event.ID),
			zap.ByteString("data", data),
			zap.Error(err))

		return
	}
}
