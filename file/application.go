package file

import (
	"context"
	"path"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

const (
	metaCreatedAtKey = "created_at"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
}

type FileApplication struct {
	repo   FileRepository
	logger *zap.Logger
}

func NewFileApplication(repo FileRepository, logger *zap.Logger) *FileApplication {
	return &FileApplication{
		repo:   repo,
		logger: logger,
	}
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
	return file, app.repo.Create(ctx, file)
}
