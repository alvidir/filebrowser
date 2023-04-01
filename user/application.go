package user

import (
	"context"
	"encoding/json"

	fb "github.com/alvidir/filebrowser"
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

// GetProfile returns the user profile instance corresponding to the given user id.
func (app *UserApplication) GetProfile(ctx context.Context, uid int32) (*Profile, error) {
	app.logger.Info("processing a \"get profile\" request",
		zap.Int32("user_id", uid))

	options := &dir.RepoOptions{
		LazyLoading: true,
	}

	dir, err := app.dirRepo.FindByUserId(ctx, uid, options)
	if err != nil {
		return nil, err
	}

	f := dir.FileByPath(profilePath)
	if f == nil {
		app.logger.Error("getting profile file from directory",
			zap.Int32("user_id", uid),
			zap.String("path", profilePath),
			zap.Error(fb.ErrNotFound))

		return nil, fb.ErrNotFound
	}

	f, err = app.fileRepo.Find(ctx, f.Id())
	if err != nil {
		return nil, err
	}

	profile := new(Profile)
	if err := json.Unmarshal(f.Data(), profile); err != nil {
		app.logger.Error("unmarshaling profile data",
			zap.Int32("user_id", uid),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return profile, nil
}
