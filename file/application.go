package file

import (
	"context"
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
	RegisterFile(ctx context.Context, uid int32, file *File) (string, error)
	UnregisterFile(ctx context.Context, uid int32, file *File) error
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

type CreateOptions struct {
	Name      string
	Directory string
	Meta      Metadata
	Data      []byte
}

func (app *FileApplication) Create(ctx context.Context, uid int32, options *CreateOptions) (*File, error) {
	app.logger.Info("processing a \"create\" file request",
		zap.String("name", options.Name),
		zap.String("directory", options.Directory),
		zap.Any("user_id", uid))

	if options.Meta == nil {
		options.Meta = make(Metadata)
	}

	options.Meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Unix(), TimestampBase)
	options.Meta[MetadataUpdatedAtKey] = options.Meta[MetadataCreatedAtKey]

	file, err := NewFile("", options.Name)
	if err != nil {
		return nil, err
	}

	file.AddPermission(uid, Owner)
	file.metadata = options.Meta
	file.data = options.Data

	if err := app.fileRepo.Create(ctx, file); err != nil {
		return nil, err
	}

	file.directory = options.Directory
	name, err := app.dirApp.RegisterFile(ctx, uid, file)
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

type UpdateOptions struct {
	Name string
	Meta Metadata
	Data []byte
}

func (app *FileApplication) Update(ctx context.Context, uid int32, fid string, options *UpdateOptions) (*File, error) {
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

	if len(options.Name) > 0 {
		file.name = options.Name
	}

	if options.Data != nil {
		file.data = options.Data
	}

	if options.Meta != nil {
		// ensure immutable data is not overwrited
		options.Meta[MetadataCreatedAtKey] = file.metadata[MetadataCreatedAtKey]
		file.metadata = options.Meta
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

	if err = app.dirApp.UnregisterFile(ctx, uid, f); err != nil {
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
