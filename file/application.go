package file

import (
	"context"
	"path"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	eb "github.com/asaskevich/EventBus"
	"go.uber.org/zap"
)

const (
	metaCreatedAtKey = "created_at"
	eventFileCreated = "file::created"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
}

type FileEventHandler interface {
	OnFileCreated(ctx context.Context, fileId, path string)
}

type FileApplication struct {
	repo   FileRepository
	bus    eb.Bus
	logger *zap.Logger
}

func NewFileApplication(repo FileRepository, logger *zap.Logger) *FileApplication {
	return &FileApplication{
		repo:   repo,
		bus:    eb.New(),
		logger: logger,
	}
}

func (app *FileApplication) Subscribe(handler FileEventHandler) error {
	return app.bus.SubscribeAsync(eventFileCreated, handler.OnFileCreated, false)
}

func (app *FileApplication) Create(ctx context.Context, fpath string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"create\" file request",
		zap.Any(fb.AuthKey, ctx.Value(fb.AuthKey)),
		zap.String("path", fpath))

	uid, err := fb.GetUid(ctx, app.logger)
	if err != nil {
		return nil, err
	}

	permissions := make(Permissions)
	permissions[uid] = Read | Write | Share | Owner

	if meta == nil {
		meta = make(Metadata)
	}

	meta[metaCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), 16)

	file := NewFile(path.Base(fpath), data, permissions, meta)
	if err := app.repo.Create(ctx, file); err != nil {
		return nil, err
	}

	app.bus.Publish(eventFileCreated, ctx, file.id, fpath)
	return file, nil
}
