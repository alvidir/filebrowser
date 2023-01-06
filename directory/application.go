package directory

import (
	"context"
	"path"
	"regexp"
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

func (app *DirectoryApplication) Retrieve(ctx context.Context, uid int32, path string, filter string) (*Directory, error) {
	app.logger.Info("processing a \"retrieve\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", path),
		zap.String("filter", filter))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	filters := make([]filterFileFn, 0, 2)
	if len(path) > 0 {
		filterFn, err := newFilterByDirFn(path)
		if err != nil {
			return nil, err
		}

		filters = append(filters, filterFn)
	}

	if len(filter) > 0 {
		filterFn, err := newFilterByRegexFn(filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, filterFn)
	}

	files, err := filterFiles(dir.Files(), filters)
	if err != nil {
		return nil, err
	}

	dir.files, err = files.Agregate()
	if err != nil {
		return nil, err
	}

	for p, f := range dir.Files() {
		f.AuthorizedFieldsOnly(uid)
		f.SetName(p)
	}

	return dir, nil
}

// Delete hard deletes the user uid directory
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

// Relocate appends the target path at the beginning of all these file paths in the directory that matches the
// filter's regex. If the regex contains more than one submatches, then the starting indexes of the first and
// second submatches determines the substring to be removed before appending.
func (app *DirectoryApplication) Relocate(ctx context.Context, uid int32, target string, filter string) error {
	app.logger.Info("processing a \"relocate\" directory request",
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

		matches := regex.FindStringSubmatchIndex(subject)
		if matches == nil {
			continue
		}

		matchStart := matches[0] + 1
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

// RemoveFiles removes, for given user, all these files whose path in the user's directory matches the path as
// a prefix and the filter as a regex. If any of both is not defined (aka. empty string) then the corresponding
// filter will be skipt.
func (app *DirectoryApplication) RemoveFiles(ctx context.Context, uid int32, path string, filter string) error {
	app.logger.Info("processing a \"remove files\" request",
		zap.Int32("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	filters := make([]filterFileFn, 0, 2)
	if len(path) > 0 {
		filterFn, err := newFilterByPrefixFn(path)
		if err != nil {
			return err
		}

		filters = append(filters, filterFn)
	}

	if len(filter) > 0 {
		filterFn, err := newFilterByRegexFn(filter)
		if err != nil {
			return err
		}

		filters = append(filters, filterFn)
	}

	files, err := filterFiles(dir.Files(), filters)
	if err != nil {
		return err
	}

	files.Range(func(p string, f *file.File) bool {
		dir.RemoveFile(f)
		err = app.dirRepo.Save(ctx, dir)
		return err == nil
	})

	return err
}

// RegisterFile registers the given file into the user uid directory. The given path may change if,
// and only if, another file with the same name exists in the same path.
func (app *DirectoryApplication) RegisterFile(ctx context.Context, file *file.File, uid int32, fpath string) (string, error) {
	app.logger.Info("processing an \"register file\" request",
		zap.Int32("user_id", uid),
		zap.String("file_id", file.Id()))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return "", err
	}

	name := dir.AddFile(file, fb.NormalizePath(fpath))
	return path.Base(name), app.dirRepo.Save(ctx, dir)
}

// UnregisterFile unregisters the given file from the directory. This action may trigger the file's
// deletion if it becomes with no owner once unregistered.
func (app *DirectoryApplication) UnregisterFile(ctx context.Context, f *file.File, uid int32) error {
	app.logger.Info("processing a \"unregister file\" request",
		zap.Int32("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	dir.RemoveFile(f)
	return app.dirRepo.Save(ctx, dir)
}
