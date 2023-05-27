package file

import (
	"context"
	"path"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
	Find(context.Context, string) (*File, error)
	FindAll(context.Context, []string) ([]*File, error)
	Save(ctx context.Context, file *File) error
	Delete(ctx context.Context, file *File) error
}

type DirectoryApplication interface {
	RegisterFile(ctx context.Context, file *File, uid int32, path string) (string, error)
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

	meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
	meta[MetadataUpdatedAtKey] = meta[MetadataCreatedAtKey]

	file, err := NewFile("", path.Base(fpath))
	if err != nil {
		return nil, err
	}

	file.AddPermission(uid, Owner)
	file.metadata = meta
	file.data = data

	if err := app.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	name, err := app.dirApp.RegisterFile(ctx, file, uid, fpath)
	if err != nil {
		return nil, err
	}

	file.SetName(name)
	return file, nil
}

func (app *FileApplication) Get(ctx context.Context, uid int32, fid string) (*File, error) {
	app.logger.Info("processing a \"get\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.fileRepo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	perm := file.Permission(uid)
	if perm&(Read|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	file.ProtectFields(uid)
	return file, nil
}

func (app *FileApplication) Update(ctx context.Context, uid int32, fid string, name string, data []byte, meta Metadata) (*File, error) {
	app.logger.Info("processing an \"update\" file request",
		zap.String("file_id", fid),
		zap.Int32("user_id", uid))

	file, err := app.fileRepo.Find(ctx, fid)
	if err != nil {
		return nil, err
	}

	if file.Permission(uid)&(Write|Owner) == 0 {
		return nil, fb.ErrNotAvailable
	}

	if len(name) > 0 {
		file.name = name
	}

	if data != nil {
		file.data = data
	}

	if meta != nil {
		// ensure immutable data is not overwrited
		meta[MetadataCreatedAtKey] = file.metadata[MetadataCreatedAtKey]
		file.metadata = meta
	}

	file.metadata[MetadataUpdatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)

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

	if err = app.dirApp.UnregisterFile(ctx, f, uid); err != nil {
		return nil, err
	}

	if f.Permission(uid)&Owner != 0 && len(f.Owners()) == 1 {
		// uid is the only owner of file f
		f.metadata[MetadataDeletedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
		err = app.fileRepo.Delete(ctx, f)
	} else if f.RevokeAccess(uid) {
		err = app.fileRepo.Save(ctx, f)
	} else {
		app.logger.Warn("unauthorized \"delete\" file request",
			zap.String("file_id", fid),
			zap.Int32("user_id", uid))
		return nil, fb.ErrNotAvailable
	}

	return f, err
}
