package user

import (
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type UserApplication struct {
	dirRepo  dir.DirectoryRepository
	fileRepo file.FileRepository
	logger   *zap.Logger
}

func NewUserApplication(dirRepo dir.DirectoryRepository, fileRepo file.FileRepository, logger *zap.Logger) *UserApplication {
	return &UserApplication{
		fileRepo: fileRepo,
		dirRepo:  dirRepo,
		logger:   logger,
	}
}

func (app *UserApplication) GetProfile(uid int32) (*UserProfile, error) {
	return nil, nil
}
