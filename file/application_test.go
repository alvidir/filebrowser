package file

import (
	"context"
	"path"
	"strconv"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

const (
	mockFileId = "000"
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

	file.id = mockFileId
	return nil
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

func TestFileApplication_create(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryAddFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		addFile: func(ctx context.Context, file *File, uid int32, path string) error {
			directoryAddFileMethodExecuted = true
			return nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, logger)

	userId := int32(999)
	fpath := "path/to/example.test"
	meta := make(Metadata)

	customFieldKey := "custom_field"
	customFieldValue := "custom value"
	meta[customFieldKey] = customFieldValue

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, fpath, meta)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if got := file.Id(); got != mockFileId {
		t.Errorf("got file.id = %v, want = %v", got, mockFileId)
	}

	if want := path.Base(fpath); want != file.Name() {
		t.Errorf("got file.name = %v, want = %v", file.name, want)
	}

	createdAt, exists := file.Value("created_at")
	if !exists {
		t.Errorf("metadata created_at does not exists")
	} else if unixCreatedAt, err := strconv.ParseInt(createdAt, 16, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixCreatedAt < before || unixCreatedAt > after {
		t.Errorf("got created_at = %v, want > %v && < %v", unixCreatedAt, before, after)
	}

	customField, exists := file.Value(customFieldKey)
	if !exists {
		t.Errorf("metadata custom_field does not exists")
	} else if customField != customFieldValue {
		t.Errorf("got custom field = %v, want = %v", customField, customFieldValue)
	}

	if !directoryAddFileMethodExecuted {
		t.Errorf("directory's AddFile method did not execute")
	}
}

func TestFileApplication_read(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "example.test",
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
