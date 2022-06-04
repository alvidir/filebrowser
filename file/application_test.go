package file

import (
	"context"
	"errors"
	"path"
	"strconv"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type directoryApplicationMock struct {
	addFile    func(ctx context.Context, file *File, uid int32, path string) error
	removeFile func(ctx context.Context, file *File, uid int32) error
}

func (app *directoryApplicationMock) AddFile(ctx context.Context, file *File, uid int32, path string) error {
	if app.addFile != nil {
		return app.addFile(ctx, file, uid, path)
	}

	return fb.ErrUnknown
}

func (app *directoryApplicationMock) RemoveFile(ctx context.Context, file *File, uid int32) error {
	if app.removeFile != nil {
		return app.removeFile(ctx, file, uid)
	}

	return fb.ErrUnknown
}

type fileRepositoryMock struct {
	create func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	find   func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error)
	save   func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	delete func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	flags  uint8
}

func (mock *fileRepositoryMock) Create(ctx context.Context, file *File) error {
	if mock.create != nil {
		return mock.create(mock, ctx, file)
	}

	return fb.ErrAlreadyExists
}

func (mock *fileRepositoryMock) Find(ctx context.Context, id string) (*File, error) {
	if mock.find != nil {
		return mock.find(mock, ctx, id)
	}

	return nil, fb.ErrNotFound
}

func (mock *fileRepositoryMock) Save(ctx context.Context, file *File) error {
	if mock.save != nil {
		return mock.save(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func (mock *fileRepositoryMock) Delete(ctx context.Context, file *File) error {
	if mock.delete != nil {
		return mock.delete(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func TestCreate(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryAddFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			directoryAddFileMethodExecuted = true
			return nil
		},
	}

	var fileId string = "999"
	fileRepo := &fileRepositoryMock{
		create: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			file.id = fileId
			return nil
		},
	}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fpath := "path/to/example.test"

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, fpath, nil)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if got := file.Id(); got != fileId {
		t.Errorf("got id = %v, want = %v", got, fileId)
	}

	if want := path.Base(fpath); want != file.Name() {
		t.Errorf("got name = %v, want = %v", file.name, want)
	}

	if createdAt, exists := file.Value("created_at"); !exists {
		t.Errorf("got created_at = %v, want > %v && < %v", createdAt, before, after)
	} else if unixCreatedAt, err := strconv.ParseInt(createdAt, tsBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixCreatedAt < before || unixCreatedAt > after {
		t.Errorf("got created_at = %v, want > %v && < %v", unixCreatedAt, before, after)
	}

	if !directoryAddFileMethodExecuted {
		t.Errorf("directory's AddFile method did not execute")
	}
}

func TestCreateWithCustomMetadata(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			return nil
		},
	}

	var fileId string = "999"
	fileRepo := &fileRepositoryMock{
		create: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			file.id = fileId
			return nil
		},
	}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fpath := "path/to/example.test"
	meta := make(Metadata)

	customFieldKey := "custom_field"
	customFieldValue := "custom value"
	meta[customFieldKey] = customFieldValue
	meta[CreatedAtKey] = strconv.FormatInt(time.Now().Add(time.Hour*24).Unix(), tsBase)

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, fpath, meta)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if createdAt, exists := file.Value("created_at"); !exists {
		t.Errorf("got created_at = %v, want > %v && < %v", createdAt, before, after)
	} else if unixCreatedAt, err := strconv.ParseInt(createdAt, tsBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixCreatedAt < before || unixCreatedAt > after {
		t.Errorf("got created_at = %v, want > %v && < %v", unixCreatedAt, before, after)
	}

	if customField, exists := file.Value(customFieldKey); !exists {
		t.Errorf("metadata custom_field does not exists")
	} else if customField != customFieldValue {
		t.Errorf("got custom field = %v, want = %v", customField, customFieldValue)
	}
}

func TestCreateWhenFileAlreadyExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			return nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fpath := "path/to/example.test"

	if _, err := app.Create(context.Background(), userId, fpath, nil); !errors.Is(err, fb.ErrAlreadyExists) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrAlreadyExists)
	}
}

func TestReadWhenFileDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			return nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Read(context.Background(), userId, fid); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestReadWhenHasNoPermissions(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			return nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Read(context.Background(), userId, fid); errors.Is(err, fb.ErrNotAvailable) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}
}

func TestRead(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: Permissions{111: Owner, 222: Read, 333: Grant | Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}

	dirApp := &directoryApplicationMock{}
	app := NewFileApplication(repo, dirApp, logger)
	file, err := app.Read(context.Background(), 111, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want := Permissions{111: Owner, 222: Read, 333: Grant | Read}
	if len(file.permissions) != len(want) {
		t.Errorf("got permissions = %+v, want = %+v", file.permissions, want)
	}

	file, err = app.Read(context.Background(), 333, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if len(file.permissions) != len(want) {
		t.Errorf("got permissions = %+v, want = %+v", file.permissions, want)
	}

	file, err = app.Read(context.Background(), 222, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want = Permissions{111: Owner, 222: Read}
	if _, exists := file.permissions[333]; exists {
		t.Errorf("got permission = %v, want = %v", file.permissions, want)
	}

	_, err = app.Read(context.Background(), 444, "")
	if err == nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
		return
	}

	repo.flags = Public
	file, err = app.Read(context.Background(), 444, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want = Permissions{111: Owner}
	if len(file.permissions) != len(want) {
		t.Errorf("got permissions = %+v, want = %+v", file.permissions, want)
	}
}
