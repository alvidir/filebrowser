package directory

import (
	"context"
	"strconv"
	"sync"
	"time"

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

func (app *DirectoryApplication) Retrieve(ctx context.Context, uid int32) (*Directory, error) {
	app.logger.Info("processing a \"retrieve\" directory request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	for _, f := range dir.Files() {
		f.HideProtectedFields(uid)
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

	var wg sync.WaitGroup
	for _, f := range dir.Files() {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, fid string) {
			defer wg.Done()

			f, err := app.fileRepo.Find(ctx, fid)
			if err != nil {
				return
			}

			if f.Permission(uid)&fb.Owner == 0 {
				return
			}

			if len(f.Owners()) == 1 {
				// uid is the only owner of file f
				f.AddMetadata(file.MetadataDeletedAtKey, strconv.FormatInt(time.Now().Unix(), file.TimestampBase))
				app.fileRepo.Delete(ctx, f)
			}
		}(ctx, &wg, f.Id())
	}

	if err := app.dirRepo.Delete(ctx, dir); err != nil {
		return err
	}

	wg.Wait()
	return nil
}

// RegisterFile is executed when a file has been created
func (app *DirectoryApplication) RegisterFile(ctx context.Context, file *file.File, uid int32, fpath string) error {
	app.logger.Info("processing an \"add file\" request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	dir.AddFile(file, fpath)
	return app.dirRepo.Save(ctx, dir)
}

// UnregisterFile is executed when a file has been deleted
func (app *DirectoryApplication) UnregisterFile(ctx context.Context, f *file.File, uid int32) error {
	app.logger.Info("processing a \"remove file\" request",
		zap.Any("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	if f.Permission(uid)&fb.Owner == 0 {
		dir.RemoveFile(f)
		return app.dirRepo.Save(ctx, dir)
	}

	if _, exists := f.Value(file.MetadataDeletedAtKey); !exists {
		dir.RemoveFile(f)
		return app.dirRepo.Save(ctx, dir)
	}

	var wg sync.WaitGroup
	for _, uid := range f.SharedWith() {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, uid int32) {
			defer wg.Done()

			dir, err := app.dirRepo.FindByUserId(ctx, uid)
			if err != nil {
				return
			}

			dir.RemoveFile(f)
			app.dirRepo.Save(ctx, dir)
		}(ctx, &wg, uid)
	}

	wg.Wait()
	return nil
}
