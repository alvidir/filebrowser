package directory

import (
	"context"

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
		zap.Int32("user_id", uid))

	directory := NewDirectory(uid)
	if err := app.repo.Create(ctx, directory); err != nil {
		return nil, err
	}

	return directory, nil
}

func (app *DirectoryApplication) Describe(ctx context.Context, uid int32) (*Directory, error) {
	app.logger.Info("processing a \"describe\" directory request",
		zap.Any("user_id", uid))

	dir, err := app.repo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	return dir, nil
}

func (app *DirectoryApplication) AddFile(ctx context.Context, file *file.File, uid int32, fpath string) error {
	app.logger.Info("processing an \"add file\" request",
		zap.Any("user_id", uid))

	dir, err := app.repo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	dir.AddFile(file, fpath)
	return app.repo.Save(ctx, dir)
}
