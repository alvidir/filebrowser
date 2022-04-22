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
	metaUpdatedAtKey = "updated_at"
	eventFileCreated = "file::created"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
	Find(context.Context, string) (*File, error)
	Save(ctx context.Context, file *File) error
}

type FileEventHandler interface {
	OnFileCreated(file *File, uid int32, path string)
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

func (app *FileApplication) publishEvent(topic string, args ...interface{}) {
	app.bus.Publish(eventFileCreated, args...)
}

func (app *FileApplication) Subscribe(handler FileEventHandler) error {
	return app.bus.SubscribeAsync(eventFileCreated, handler.OnFileCreated, false)
}

func (app *FileApplication) Create(ctx context.Context, uid int32, fpath string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"create\" file request",
		zap.String("file_path", fpath),
		zap.Any("user_id", uid))

	if meta == nil {
		meta = make(Metadata)
	}

	meta[metaCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), 16)
	meta[metaUpdatedAtKey] = meta[metaCreatedAtKey]

	file := NewFile("", path.Base(fpath), data)
	file.metadata = meta

	if err := app.repo.Create(ctx, file); err != nil {
		return nil, err
	}

	file.AddPermissions(uid, Read|Write|Grant|Owner)
	app.publishEvent(eventFileCreated, file, uid, fpath)
	return file, nil
}

func (app *FileApplication) Read(ctx context.Context, uid int32, fid string) (*File, error) {
	app.logger.Info("processing a \"read\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.repo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	perm := file.Permissions(uid)
	if file.flags&Public == 0 && perm&(Read|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	if perm&Owner > 0 || perm&Grant > 0 {
		return file, nil
	}

	for id, p := range file.permissions {
		if id != uid && p&Owner == 0 {
			delete(file.permissions, id)
		}
	}

	return file, nil
}

func (app *FileApplication) Write(ctx context.Context, uid int32, fid string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"write\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.repo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	if file.Permissions(uid)&(Write|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	file.data = data
	if meta != nil {
		meta[metaCreatedAtKey] = file.metadata[metaCreatedAtKey]
		file.metadata = meta
	}

	file.metadata[metaUpdatedAtKey] = strconv.FormatInt(time.Now().Unix(), 16)

	if err := app.repo.Save(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}
