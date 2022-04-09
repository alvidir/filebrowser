package directory

import (
	"context"
	"errors"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type DirectoryRepository interface {
	FindByUserId(ctx context.Context, userId int32) (*Directory, error)
	Create(ctx context.Context, directory *Directory) error
}

type DirectoryApplication struct {
	directoryRepo DirectoryRepository
	logger        *zap.Logger
}

func NewDirectoryApplication(repo DirectoryRepository, logger *zap.Logger) *DirectoryApplication {
	return &DirectoryApplication{
		directoryRepo: repo,
		logger:        logger,
	}
}

func (app *DirectoryApplication) Create(ctx context.Context, userId int32) (*Directory, error) {
	app.logger.Info("processing a \"create\" directory request",
		zap.Int32("user", userId))

	if _, err := app.directoryRepo.FindByUserId(ctx, userId); err == nil {
		return nil, fb.ErrAlreadyExists
	} else if !errors.Is(err, fb.ErrNotFound) {
		return nil, err
	}

	directory := NewDirectory(userId)
	if err := app.directoryRepo.Create(ctx, directory); err != nil {
		return nil, err
	}

	return directory, nil
}
