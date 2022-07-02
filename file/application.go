package file

import (
	"context"
	"path"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

const (
	CreatedAtKey  = "created_at"
	UpdatedAtKey  = "updated_at"
	DeletedAtKey  = "deleted_at"
	TimestampBase = 16
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
	Find(context.Context, string) (*File, error)
	Save(ctx context.Context, file *File) error
	Delete(ctx context.Context, file *File) error
}

type DirectoryApplication interface {
	FileSearch(ctx context.Context, uid int32, search string) ([]*File, error)
	RegisterFile(ctx context.Context, file *File, uid int32, path string) error
	UnregisterFile(ctx context.Context, file *File, uid int32) error
}

type FileApplication struct {
	fileRepo FileRepository
	dirApp   DirectoryApplication
	logger   *zap.Logger
}

func NewFileApplication(repo FileRepository, dirApp DirectoryApplication, logger *zap.Logger) *FileApplication {
	return &FileApplication{
		fileRepo: repo,
		dirApp:   dirApp,
		logger:   logger,
	}
}

func (app *FileApplication) Create(ctx context.Context, uid int32, fpath string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"create\" file request",
		zap.String("file_path", fpath),
		zap.Any("user_id", uid))

	if meta == nil {
		meta = make(Metadata)
	}

	meta[CreatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
	meta[UpdatedAtKey] = meta[CreatedAtKey]

	file, err := NewFile("", path.Base(fpath))
	if err != nil {
		return nil, err
	}

	file.AddPermissions(uid, Read|Write|Grant|Owner)
	file.metadata = meta
	file.data = data

	if err := app.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	err = app.dirApp.RegisterFile(ctx, file, uid, fpath)
	return file, err
}

func (app *FileApplication) Read(ctx context.Context, uid int32, fid string) (*File, error) {
	app.logger.Info("processing a \"read\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.fileRepo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	perm := file.Permissions(uid)
	if file.flags&Public == 0 && perm&(Read|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	if perm&(Owner|Grant) > 0 {
		return file, nil
	}

	for id, p := range file.permissions {
		// hide all those permissions that do not belong to any of both, the user or owners
		// WARNING: DO NOT SAVE THE FOLLOWING FILE CHANGES
		if id != uid && p&Owner == 0 {
			delete(file.permissions, id)
		}
	}

	return file, nil
}

func (app *FileApplication) Write(ctx context.Context, uid int32, fid string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing a \"write\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.fileRepo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	if file.Permissions(uid)&(Write|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	file.data = data
	if meta != nil {
		// ensure immutable data is not overwrited
		meta[CreatedAtKey] = file.metadata[CreatedAtKey]
		file.metadata = meta
	}

	file.metadata[UpdatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)

	if err := app.fileRepo.Save(ctx, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (app *FileApplication) Delete(ctx context.Context, uid int32, fid string) (*File, error) {
	app.logger.Info("processing a \"delete\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	f, err := app.fileRepo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	if f.Permissions(uid)&Owner != 0 && len(f.Owners()) == 1 {
		// uid is the only owner of file f
		f.metadata[DeletedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
		err = app.fileRepo.Delete(ctx, f)
	} else if f.RevokeAccess(uid) {
		err = app.fileRepo.Save(ctx, f)
	} else {
		app.logger.Warn("unauthorized \"delete\" file request",
			zap.String("file_id", fid),
			zap.Int32("user_id", uid))
		return nil, fb.ErrNotAvailable
	}

	if err != nil {
		return nil, err
	}

	err = app.dirApp.UnregisterFile(ctx, f, uid)
	return f, err
}

func (app *FileApplication) Search(ctx context.Context, uid int32, search string) ([]*File, error) {
	app.logger.Info("processing a \"search\" file request",
		zap.String("search", search),
		zap.Int32("user_id", uid))

	return app.dirApp.FileSearch(ctx, uid, search)
}
