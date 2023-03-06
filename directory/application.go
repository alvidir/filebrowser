package directory

import (
	"context"
	"path"
	"path/filepath"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
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

// Get agregates into a list of files all those files matching the given path.
func (app *DirectoryApplication) Get(ctx context.Context, uid int32, p string) (*Directory, error) {
	app.logger.Info("processing a \"get\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", p))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	absP := filepath.Join(PathSeparator, p)

	selected := NewDirectory(uid)
	selected.files = dir.AggregateFiles(absP)
	selected.path = absP
	selected.id = dir.id

	for _, f := range selected.files {
		f.ProtectFields(uid)
	}

	return selected, nil
}

// Delete removes from the directory all those files whose path matches the given one.
func (app *DirectoryApplication) Delete(ctx context.Context, uid int32, p string) (*Directory, error) {
	app.logger.Info("processing a \"delete\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", p))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	absP := filepath.Join(PathSeparator, p)
	affected := NewDirectory(uid)
	affected.files = dir.FilesByPath(absP)
	affected.path = absP

	for _, f := range affected.files {
		dir.RemoveFile(f)
		if f.Permission(uid)&cert.Owner == 0 {
			continue
		}

		if len(f.Owners()) > 1 {
			continue
		}

		f.AddMetadata(file.MetadataDeletedAtKey, strconv.FormatInt(time.Now().Unix(), file.TimestampBase))
		if err := app.fileRepo.Delete(ctx, f); err != nil {
			return nil, err
		}

		f.ProtectFields(uid)
	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return nil, err
	}

	for _, f := range affected.files {
		f.ProtectFields(uid)
	}

	return affected, nil
}

// Move replaces the destination path to all these file paths in the directory matching any of the given paths.
func (app *DirectoryApplication) Move(ctx context.Context, uid int32, paths []string, dest string) (*Directory, error) {
	app.logger.Info("processing a directory's \"move\" request",
		zap.Int32("user_id", uid),
		zap.Strings("paths", paths),
		zap.String("destination", dest))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	destDir := path.Dir(dest)
	absDest := filepath.Join(PathSeparator, dest)

	affected := NewDirectory(uid)
	prefixes := make(map[string]string)
	for _, p := range paths {
		absP := filepath.Join(PathSeparator, p)
		for absFp, f := range dir.FilesByPath(absP) {
			affected.files[absFp] = f
			prefixes[absFp] = absP
		}
	}

	for absFp, f := range affected.files {
		sufix := absFp[len(prefixes[absFp]):]
		finalPath := path.Join(absDest, sufix)
		if destDir == absDest {
			finalPath = path.Join(absDest, path.Base(absFp))
		}

		dir.RemoveFile(f)
		dir.AddFile(f, finalPath)

		delete(affected.files, absFp)
		affected.files[finalPath] = f
	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return nil, err
	}

	for _, f := range affected.files {
		f.ProtectFields(uid)
	}

	return affected, nil
}

// RegisterFile registers the given file into the user uid directory. The given path may change if,
// and only if, another file with the same name exists in the same path.
func (app *DirectoryApplication) RegisterFile(ctx context.Context, file *file.File, uid int32, fp string) (string, error) {
	app.logger.Info("processing a directory's \"register file\" request",
		zap.Int32("user_id", uid),
		zap.String("file_id", file.Id()),
		zap.String("path", fp))

	absFp := filepath.Join(PathSeparator, fp)
	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return "", err
	}

	name := dir.AddFile(file, absFp)
	return path.Base(name), app.dirRepo.Save(ctx, dir)
}

// UnregisterFile unregisters the given file from the directory. This action may trigger the file's
// deletion if it becomes with no owner once unregistered.
func (app *DirectoryApplication) UnregisterFile(ctx context.Context, f *file.File, uid int32) error {
	app.logger.Info("processing a directory's \"unregister file\" request",
		zap.Int32("user_id", uid))

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return err
	}

	dir.RemoveFile(f)
	return app.dirRepo.Save(ctx, dir)
}
