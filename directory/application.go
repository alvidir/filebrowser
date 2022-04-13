package directory

import (
	"context"
	"path"

	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type DirectoryRepository interface {
	FindByUserId(ctx context.Context, userId int32) (*Directory, error)
	Create(ctx context.Context, directory *Directory) error
	Save(ctx context.Context, directory *Directory) error
}

type DirectoryApplication struct {
	repo   DirectoryRepository
	logger *zap.Logger
}

func NewDirectoryApplication(repo DirectoryRepository, logger *zap.Logger) *DirectoryApplication {
	return &DirectoryApplication{
		repo:   repo,
		logger: logger,
	}
}

func (app *DirectoryApplication) Create(ctx context.Context, uid int32) (*Directory, error) {
	app.logger.Info("processing a \"create\" directory request",
		zap.Any("uid", uid))

	directory := NewDirectory(uid)
	if err := app.repo.Create(ctx, directory); err != nil {
		return nil, err
	}

	return directory, nil
}

func (app *DirectoryApplication) AddFile(ctx context.Context, uid int32, fileId, fpath string, shared bool) error {
	app.logger.Info("processing an \"add file\" request",
		zap.Any("uid", uid))

	dir, err := app.repo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	base := path.Base(fpath)
	file := file.NewFile(fileId, base, nil, nil, nil)
	dir.AddFile(file, fpath)

	return app.repo.Save(ctx, dir)
}
