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
		if regex.MatchString(f.Name()) {
			return p, f
		}

		return "", nil
	}

	return filterFn, nil
}

func NewFilterByDirFn(target string) (FilterFileFn, error) {
	if !path.IsAbs(target) {
		target = path.Join(PathSeparator, target)
	}

	if target == PathSeparator {
		target = ""
	}

	depth := len(strings.Split(target, PathSeparator))
	filterFn := func(p string, f *file.File) (string, *file.File) {
		if !path.IsAbs(p) {
			p = path.Join(PathSeparator, p)
		}

		if strings.Compare(p, target) == 0 {
			// 0 means p == target, so is not filtering by path, but by a filename
			return "", nil
		} else if !strings.HasPrefix(p, target) {
			return "", nil
		}

		items := strings.Split(p, PathSeparator)
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

func getPathsDivergenceIndex(p1 string, p2 string) (index int) {
	p1Split := strings.Split(p1, PathSeparator)
	p2Split := strings.Split(p2, PathSeparator)

	for index = 0; index < len(p1Split) && index < len(p2Split); index++ {
		if p1Split[index] != p2Split[index] {
			break
		}
	}

	return
}

func (app *DirectoryApplication) Relocate(ctx context.Context, uid int32, target string, filter string) error {
	app.logger.Info("processing a \"move\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", target),
		zap.String("filter", filter))

	if len(target) == 0 {
		return fb.ErrNotFound
	}

	regex, err := regexp.Compile(filter)
	if err != nil {
		return fb.ErrInvalidFormat
	}

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	matches := 0
	target = path.Clean(target)
	for subject := range dir.Files() {
		subject = path.Clean(subject)
		if strings.HasSuffix(subject, target) || !regex.MatchString(subject) {
			continue
		}

		index := getPathsDivergenceIndex(target, subject)
		relative := path.Join(strings.Split(subject, PathSeparator)[index:]...)
		p := path.Join(target, relative)

		dirs := strings.Split(p, PathSeparator)
		for index := 0; index < len(dirs); index++ {
			pp := path.Join(dirs[0 : len(dirs)-index]...)
			if _, exists := dir.Files()[pp]; exists {
				return fb.ErrAlreadyExists
			}
		}

		if _, exists := dir.Files()[p]; exists {
			return fb.ErrAlreadyExists
		}

		dir.Files()[p] = dir.Files()[subject]
		delete(dir.Files(), subject)
		matches++
	}

	if matches == 0 {
		return fb.ErrNotFound
	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return err
	}

	return nil
}

// RegisterFile is executed when a file has been created
func (app *DirectoryApplication) RegisterFile(ctx context.Context, file *file.File, uid int32, fpath string) error {
	app.logger.Info("processing an \"add file\" request",
		zap.Int32("user_id", uid))

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
