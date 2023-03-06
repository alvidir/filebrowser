package directory

import (
	"context"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

func (app *DirectoryApplication) Get(ctx context.Context, uid int32, p string) (*Directory, error) {
	app.logger.Info("processing a \"get\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", p))

	pDir := path.Dir(p)
	absP := filepath.Join(PathSeparator, p)

	prefix := absP
	if absP == pDir && absP != PathSeparator {
		prefix = absP + PathSeparator
	}

	pCount := strings.Count(p, PathSeparator)
	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	selected := NewDirectory(uid)
	selected.id = dir.id
	selected.path = absP

	folders := make(map[string]int)

	for fp, f := range dir.Files() {
		absFp := filepath.Join(PathSeparator, fp)
		fpCount := strings.Count(absFp, PathSeparator)
		isSelected := false
		if absP == pDir {
			isSelected = strings.HasPrefix(absFp, prefix)
		} else {
			isSelected = absFp == absP
		}

		if !isSelected {
			continue
		}

		if pCount < fpCount {
			folderPath := filepath.Join(pathComponents(absFp)[0 : pCount+1]...)
			folderSize, exists := folders[folderPath]
			if !exists {
				folderSize = 0
			}

			folders[folderPath] = folderSize + 1
			continue
		}

		f.ProtectFields(uid)
		f.SetDirectory(absP)
		f.SetName(path.Base(absFp))

		selected.files[absFp] = f
	}

	for folderPath, folderSize := range folders {
		folder, err := file.NewFile("", path.Base(folderPath))
		if err != nil {
			continue
		}

		folder.SetFlag(file.Directory)
		folder.SetDirectory(path.Dir(folderPath))
		folder.AddMetadata(file.MetadataSizeKey, strconv.Itoa(folderSize))
		selected.files[folderPath] = folder
	}

	return selected, nil
}

// Delete removes from the directory all those files whose path matches any of the given ones.
func (app *DirectoryApplication) Delete(ctx context.Context, uid int32, p string) (*Directory, error) {
	app.logger.Info("processing a \"delete\" directory request",
		zap.Int32("user_id", uid),
		zap.String("path", p))

	pDir := path.Dir(p)
	absP := filepath.Join(PathSeparator, p)

	prefix := absP
	if absP == pDir && absP != PathSeparator {
		prefix = absP + PathSeparator
	}

	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	for fp, f := range dir.Files() {
		absFp := filepath.Join(PathSeparator, fp)
		isSelected := false
		if absP == pDir {
			isSelected = strings.HasPrefix(absFp, prefix)
		} else {
			isSelected = absFp == absP
		}

		if !isSelected {
			continue
		}

		delete(dir.files, fp)

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

	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return nil, err
	}

	dir.path = absP
	return dir, nil
}

// Move replaces the destination path to all these file paths in the directory matching any of the given paths.
func (app *DirectoryApplication) Move(ctx context.Context, uid int32, paths []string, dest string) (*Directory, error) {
	app.logger.Info("processing a directory's \"move\" request",
		zap.Int32("user_id", uid),
		zap.Strings("paths", paths),
		zap.String("destination", dest))

	destDir := path.Dir(dest)
	absDest := filepath.Join(PathSeparator, dest)
	dir, err := app.dirRepo.FindByUserId(ctx, uid)
	if err != nil {
		return nil, err
	}

	for fp, f := range dir.Files() {
		absFp := filepath.Join(PathSeparator, fp)

		isSelected := false
		var prefix string
		for _, p := range paths {
			pDir := path.Dir(p)
			absP := filepath.Join(PathSeparator, p)

			prefix = absP
			if absP == pDir && absP != PathSeparator {
				prefix = absP + PathSeparator
			}

			if absP == pDir {
				isSelected = strings.HasPrefix(absFp, prefix)
			} else {
				isSelected = absFp == absP
			}

			if isSelected {
				break
			}
		}

		if !isSelected {
			continue
		}

		sufix := absFp[len(prefix):]
		finalPath := path.Join(absDest, sufix)
		if destDir == absDest {
			finalPath = path.Join(absDest, path.Base(absFp))
		}

		delete(dir.files, fp)
		dir.AddFile(f, finalPath)
	}

	if err := app.dirRepo.Save(ctx, dir); err != nil {
		return nil, err
	}

	// TODO: send only dest directory
	return dir, nil
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
