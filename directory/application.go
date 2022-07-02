package directory

import (
	"context"
	"regexp"
	"sort"
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
	for _, fileId := range dir.Files() {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, fid string) {
			defer wg.Done()

			f, err := app.fileRepo.Find(ctx, fid)
			if err != nil {
				return
			}

			if f.Permissions(uid)&file.Owner == 0 {
				return
			}

			if len(f.Owners()) == 1 {
				// uid is the only owner of file f
				f.AddValue(file.DeletedAtKey, strconv.FormatInt(time.Now().Unix(), file.TimestampBase))
				app.fileRepo.Delete(ctx, f)
			}
		}(ctx, &wg, fileId)
	}

	if err := app.dirRepo.Delete(ctx, dir); err != nil {
		return err
	}

	wg.Wait()
	return nil
}

func (app *DirectoryApplication) FileSearch(ctx context.Context, uid int32, search string) ([]*file.File, error) {
	app.logger.Info("processing a \"search file\" directory request",
		zap.Any("user_id", uid),
		zap.String("search", search))

	dir, err := app.Retrieve(ctx, uid)
	if err != nil {
		return nil, err
	}

	regex, err := regexp.Compile(search)
	if err != nil {
		app.logger.Error("compiling search regex",
			zap.Any("user_id", uid),
			zap.String("search", search),
			zap.Error(err))
	}

	length := 0
	filenames := make([]string, len(dir.files))

	for path := range dir.files {
		if regex.FindStringIndex(path) == nil {
			continue
		}

		filenames[length] = path
		length++
	}

	filenames = filenames[:length]
	sort.Slice(filenames, func(i, j int) bool {
		return regex.FindStringIndex(filenames[i])[0] <
			regex.FindStringIndex(filenames[j])[0]
	})

	files := make([]*file.File, length)
	for index, filename := range filenames {
		files[index], _ = file.NewFile(dir.files[filename], filename)
	}

	return files, nil
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

	if f.Permissions(uid)&file.Owner == 0 {
		dir.RemoveFile(f)
		return app.dirRepo.Save(ctx, dir)
	}

	if _, exists := f.Value(file.DeletedAtKey); !exists {
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
