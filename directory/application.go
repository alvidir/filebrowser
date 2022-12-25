package directory

import (
	"context"
	"path"
	"regexp"
	"strconv"
	"strings"
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

func NewFilterByNameFn(target string) (FilterFileFn, error) {
	regex, err := regexp.Compile(target)
	if err != nil {
		return nil, fb.ErrInvalidFormat
	}

	filterFn := func(p string, f *file.File) (string, *file.File) {
		if regex.MatchString(p) || regex.MatchString(f.Name()) {
			return p, f
		}

		return "", nil
	}

	return filterFn, nil
}

func NewFilterByDirFn(target string) (FilterFileFn, error) {
	target = fb.NormalizePath(target)

	depth := len(fb.PathComponents(target))
	filterFn := func(p string, f *file.File) (string, *file.File) {
		p = fb.NormalizePath(p)

		if strings.Compare(p, target) == 0 {
			// 0 means p == target, so is not filtering by path, but by a filename
			return "", nil
		} else if !strings.HasPrefix(p, target) {
			return "", nil
		}

		items := fb.PathComponents(p)
		name := items[depth]

		if len(items) > depth+1 {
			f, _ = file.NewFile("", name)
			f.SetFlag(file.Directory)
		}

		return name, f
	}

	return filterFn, nil
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

func (app *DirectoryApplication) Retrieve(ctx context.Context, uid int32, path string, filter string) (*Directory, error) {
	app.logger.Info("processing a \"retrieve\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", path),
		zap.String("filter", filter))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	filters := make([]FilterFileFn, 0, 2)
	if len(path) > 0 {
		filterFn, err := NewFilterByDirFn(path)
		if err != nil {
			return nil, err
		}

		filters = append(filters, filterFn)
	}

	if len(filter) > 0 {
		filterFn, err := NewFilterByNameFn(filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, filterFn)
	}

	dir.files, err = dir.FilterFiles(filters)
	if err != nil {
		return nil, err
	}

	for p, f := range dir.Files() {
		f.AuthorizedFieldsOnly(uid)
		f.SetName(p)
	}

	return dir, nil
}

func (app *DirectoryApplication) Delete(ctx context.Context, uid int32) error {
	app.logger.Info("processing a \"delete\" directory request",
		zap.Int32("user_id", uid))

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

func (app *DirectoryApplication) Relocate(ctx context.Context, uid int32, target string, filter string) error {
	app.logger.Info("processing a \"move\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", target),
		zap.String("filter", filter))

	regex, err := regexp.Compile(filter)
	if err != nil {
		return fb.ErrInvalidFormat
	}

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	target = fb.NormalizePath(target)
	for p0, f := range dir.Files() {
		subject := fb.NormalizePath(p0)

		matches := regex.FindStringIndex(subject)
		if matches == nil {
			continue
		}

		matchStart := matches[0] + 1
		matches = regex.FindStringSubmatchIndex(subject)
		if regex.NumSubexp() > 0 && matches[2] > -1 {
			matchStart = matches[2]
		}

		index := len(fb.PathComponents(subject[:matchStart]))
		p1 := path.Join(fb.PathComponents(subject)[index:]...)
		p1 = fb.NormalizePath(path.Join(target, p1))

		delete(dir.files, p0)
		dir.AddFile(f, p1)
	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return err
	}

	return nil
}

// RegisterFile is executed when a file has been created
func (app *DirectoryApplication) RegisterFile(ctx context.Context, file *file.File, uid int32, fpath string) (string, error) {
	app.logger.Info("processing an \"add file\" request",
		zap.Int32("user_id", uid),
		zap.String("file_id", file.Id()))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return "", err
	}

	name := dir.AddFile(file, fb.NormalizePath(fpath))
	return path.Base(name), app.dirRepo.Save(ctx, dir)
}

// UnregisterFile is executed when a file has been deleted
func (app *DirectoryApplication) UnregisterFile(ctx context.Context, f *file.File, uid int32) error {
	app.logger.Info("processing a \"remove file\" request",
		zap.Int32("user_id", uid))

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
