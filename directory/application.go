package directory

import (
	"context"
	"errors"
	"fmt"

	fb "github.com/alvidir/filebrowser"
)

var (
	ErrAlreadyExists = errors.New("already exists")
)

type DirectoryRepository interface {
	FindByUserId(ctx context.Context, userId int32) (*Directory, error)
	Create(ctx context.Context, directory *Directory) error
}

type DirectoryApplication struct {
	DirectoryRepo DirectoryRepository
	Logger        fb.Logger
}

func (app *DirectoryApplication) Create(ctx context.Context, userId int32) error {
	app.Logger.Infof("processing a \"create\" directory request for user %s", userId)

	if _, err := app.DirectoryRepo.FindByUserId(ctx, userId); err == nil {
		return fmt.Errorf("directory with user id %d: %w", userId, ErrAlreadyExists)
	}

	directory := NewDirectory(userId)
	if err := app.DirectoryRepo.Create(ctx, directory); err != nil {
		app.Logger.Errorf("creating directory for user %s: %s", userId, err)
		return err
	}

	return nil
}
