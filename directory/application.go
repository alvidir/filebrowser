package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type DirectoryRepository interface {
	FindByUserId(ctx context.Context, userId int32) (*Directory, error)
	Create(ctx context.Context, directory *Directory) error
	Save(ctx context.Context, directory *Directory) error
	Delete(ctx context.Context, directory *Directory) error
}

type DirectoryApplication struct {
	dirRepo  DirectoryRepository
	fileRepo file.FileRepository
	logger   *zap.Logger
}

func NewDirectoryApplication(dirRepo DirectoryRepository, fileRepo file.FileRepository, logger *zap.Logger) *DirectoryApplication {
	return &DirectoryApplication{
		dirRepo:  dirRepo,
		fileRepo: fileRepo,
		logger:   logger,
	}
}

func (app *DirectoryApplication) Create(ctx context.Context, uid int32) (*Directory, error) {
	app.logger.Info("processing a \"create\" directory request",
		zap.Int32("user_id", uid))

	if _, err := app.dirRepo.FindByUserId(ctx, uid); err == nil {
		return nil, fb.ErrAlreadyExists
	}

	directory := NewDirectory(uid)
	if err := app.dirRepo.Create(ctx, directory); err != nil {
		return nil, err
	}

	return directory, nil
}

func (app *DirectoryApplication) Describe(ctx context.Context, uid int32) (*Directory, error) {
	app.logger.Info("processing a \"describe\" directory request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	return dir, nil
}

func (app *DirectoryApplication) Delete(ctx context.Context, uid int32) error {
	app.logger.Info("processing a \"delete\" directory request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	for _, fileId := range dir.List() {
		f, err := app.fileRepo.Find(ctx, fileId)
		if err != nil {
			continue
		}

		if f.Permissions(uid)&file.Owner == 0 {
			continue
		}

		// TODO: the following condition requires the file to be unique and lockable
		// for the whole system
		if owners := f.Owners(); len(owners) > 1 {
			app.logger.Warn("unsafe condition evaluated",
				zap.String("reason", "the File instance has to be unique and lockable"))

			f.RevokeAccess(dir.userId)
			err = app.fileRepo.Save(ctx, f)
		} else {
			err = app.fileRepo.Delete(ctx, f)
		}

		if err != nil {
			return err
		}
	}

	if err := app.dirRepo.Delete(ctx, dir); err != nil {
		return err
	}

	return nil
}

// AddFile is executed when a file is created and must be added into the owner's directory file list
func (app *DirectoryApplication) AddFile(ctx context.Context, file *file.File, uid int32, fpath string) error {
	app.logger.Info("processing an \"add file\" request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	dir.AddFile(file, fpath)
	return app.dirRepo.Save(ctx, dir)
}

// RemoveFile is executed when a file is deleted and must be removed from the directories file list
func (app *DirectoryApplication) RemoveFile(ctx context.Context, f *file.File, owner int32) error {
	app.logger.Info("processing a \"remove file\" request",
		zap.Any("user_id", owner))

	dir, err := app.dirRepo.FindByUserId(ctx, owner)
	if err != nil {
		return err
	}

	if _, exists := f.Value(file.DeletedAtKey); !exists {
		dir.RemoveFile(f)
		return app.dirRepo.Save(ctx, dir)
	}

	// TODO: the file must be deleted from all directories file list
	return nil
}
