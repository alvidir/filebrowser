package file

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type directoryApplicationMock struct {
	registerFile   func(ctx context.Context, uid int32, file *File) (string, error)
	unregisterFile func(ctx context.Context, uid int32, file *File) error
}

func (app *directoryApplicationMock) RegisterFile(ctx context.Context, uid int32, file *File) (string, error) {
	if app.registerFile != nil {
		return app.registerFile(ctx, uid, file)
	}

	return "", fb.ErrUnknown
}

func (app *directoryApplicationMock) UnregisterFile(ctx context.Context, uid int32, file *File) error {
	if app.unregisterFile != nil {
		return app.unregisterFile(ctx, uid, file)
	}

	return fb.ErrUnknown
}

func (app *directoryApplicationMock) FileSearch(ctx context.Context, uid int32, search string) ([]*File, error) {
	return nil, fb.ErrUnknown
}

type EventBusMock struct {
	emitFileCreated func(repo *EventBusMock, uid int32, f *File) error
	emitFileDeleted func(repo *EventBusMock, uid int32, f *File) error
}

func (bus *EventBusMock) EmitFileCreated(uid int32, f *File) error {
	if bus.emitFileCreated != nil {
		return bus.emitFileCreated(bus, uid, f)
	}

	return nil
}

func (bus *EventBusMock) EmitFileDeleted(uid int32, f *File) error {
	if bus.emitFileDeleted != nil {
		return bus.emitFileDeleted(bus, uid, f)
	}

	return nil
}

type fileRepositoryMock struct {
	create func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	find   func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error)
	save   func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	delete func(repo *fileRepositoryMock, ctx context.Context, file *File) error
	flags  Flag
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

func (mock *fileRepositoryMock) FindAll(context.Context, []string) ([]*File, error) {
	return nil, errors.New("unimplemented")
}

func (mock *fileRepositoryMock) FindPermissions(context.Context, string) (map[int32]Permission, error) {
	return nil, errors.New("unimplemented")
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

func TestCreateWhenFileAlreadyExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	options := CreateOptions{
		Name:      "example.test",
		Directory: "path/to",
	}

	if _, err := app.Create(context.Background(), userId, &options); !errors.Is(err, fb.ErrAlreadyExists) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrAlreadyExists)
	}
}

func TestReadWhenFileDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Get(context.Background(), userId, fid); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestCreate(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryAddFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			directoryAddFileMethodExecuted = true
			return file.name, nil
		},
	}

	var fileId string = "999"
	fileRepo := &fileRepositoryMock{
		create: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			file.id = fileId
			return nil
		},
	}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	options := CreateOptions{
		Name:      "example.test",
		Directory: "path/to",
		Data:      []byte("hello world"),
	}

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, &options)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if got := file.Id(); got != fileId {
		t.Errorf("got id = %v, want = %v", got, fileId)
	}

	if want := options.Name; want != file.Name() {
		t.Errorf("got name = %v, want = %v", file.name, want)
	}

	if got := file.Data(); string(got) != string(options.Data) {
		t.Errorf("got data = %v, want = %v", got, options.Data)
	}

	if createdAt, exists := file.metadata[MetadataCreatedAtKey]; !exists {
		t.Errorf("got created_at = %v, want > %v && < %v", createdAt, before, after)
	} else if unixCreatedAt, err := strconv.ParseInt(createdAt, TimestampBase, 64); err != nil {
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
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	var fileId string = "999"
	fileRepo := &fileRepositoryMock{
		create: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			file.id = fileId
			return nil
		},
	}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	options := CreateOptions{
		Name:      "example.test",
		Directory: "path/to",
		Meta:      make(Metadata),
	}

	customFieldKey := "custom_field"
	customFieldValue := "custom value"
	options.Meta[customFieldKey] = customFieldValue
	options.Meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Add(time.Hour*24).Unix(), TimestampBase)

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, &options)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	if createdAt, exists := file.metadata[MetadataCreatedAtKey]; !exists {
		t.Errorf("got created_at = %v, want > %v && < %v", createdAt, before, after)
	} else if unixCreatedAt, err := strconv.ParseInt(createdAt, TimestampBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixCreatedAt < before || unixCreatedAt > after {
		t.Errorf("got created_at = %v, want > %v && < %v", unixCreatedAt, before, after)
	}

	if customField, exists := file.metadata[customFieldKey]; !exists || customField != customFieldValue {
		t.Errorf("got custom_field = %v, want = %v", customField, customFieldValue)
	}
}

func TestReadWhenHasNoPermissions(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Get(context.Background(), userId, fid); errors.Is(err, fb.ErrNotAvailable) {
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
				permissions: map[int32]Permission{111: Owner, 222: Read, 333: Write | Read, 444: Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}

	dirApp := &directoryApplicationMock{}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)
	file, err := app.Get(context.Background(), 111, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want := map[int32]Permission{111: Owner, 222: Read, 333: Write | Read, 444: Read}
	if len(file.permissions) != len(want) {
		t.Errorf("got permissions = %+v, want = %+v", file.permissions, want)
	}

	file, err = app.Get(context.Background(), 333, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want = map[int32]Permission{111: Owner, 222: Read, 333: Write | Read, 444: Read}
	if len(file.permissions) != len(want) {
		t.Errorf("got permissions = %+v, want = %+v", file.permissions, want)
	}

	file, err = app.Get(context.Background(), 222, "")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want = map[int32]Permission{111: Owner, 222: Read, 333: Write | Read}
	if _, exists := file.permissions[444]; exists {
		t.Errorf("got permission = %v, want = %v", file.permissions, want)
	}

	_, err = app.Get(context.Background(), 555, "")
	if !errors.Is(err, fb.ErrNotAvailable) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
		return
	}
}

func TestWriteWhenFileDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Update(context.Background(), userId, fid, &UpdateOptions{}); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestWriteWhenHasNoPermissions(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{111: Owner, 222: Read, 333: Write | Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	if _, err := app.Update(context.Background(), 222, fid, &UpdateOptions{}); !errors.Is(err, fb.ErrNotAvailable) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}

	if _, err := app.Update(context.Background(), 999, fid, &UpdateOptions{}); !errors.Is(err, fb.ErrNotAvailable) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}
}

func TestWriteWhenCannotSave(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{111: Owner, 222: Read, 333: Write | Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}

	dirApp := &directoryApplicationMock{}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	if _, err := app.Update(context.Background(), 111, fid, &UpdateOptions{}); !errors.Is(err, fb.ErrUnknown) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrUnknown)
	}
}

func TestWrite(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	meta := make(Metadata)
	createdAtValue := "000"
	meta[MetadataCreatedAtKey] = createdAtValue

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    meta,
				permissions: map[int32]Permission{111: Owner, 222: Read, 333: Write | Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},

		save: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			return nil
		},
	}

	dirApp := &directoryApplicationMock{}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	options := UpdateOptions{
		Data: []byte{1, 2, 3},
	}

	before := time.Now().Unix()
	file, err := app.Update(context.Background(), 111, fid, &options)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}

	if createdAt, exists := file.metadata[MetadataCreatedAtKey]; !exists || createdAt != createdAtValue {
		t.Errorf("got created_at = %v, want = %v", createdAt, createdAtValue)
	}

	if updatedAt, exists := file.metadata[MetadataUpdatedAtKey]; !exists {
		t.Errorf("got updated_at = %v, want > %v && < %v", updatedAt, before, after)
	} else if unixUpdatedAt, err := strconv.ParseInt(updatedAt, TimestampBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixUpdatedAt < before || unixUpdatedAt > after {
		t.Errorf("got updated_at = %v, want > %v && < %v", unixUpdatedAt, before, after)
	}
}

func TestWriteWithCustomMetadata(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	meta := make(Metadata)
	createdAtValue := "000"
	meta[MetadataCreatedAtKey] = createdAtValue

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    meta,
				permissions: map[int32]Permission{111: Owner, 222: Read, 333: Write | Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},

		save: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			return nil
		},
	}

	dirApp := &directoryApplicationMock{}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	options := UpdateOptions{
		Meta: make(Metadata),
		Data: []byte{1, 2, 3},
	}

	customFieldKey := "custom_field"
	customFieldValue := "custom value"
	options.Meta[customFieldKey] = customFieldValue
	options.Meta[MetadataCreatedAtKey] = strconv.FormatInt(time.Now().Add(time.Hour*24).Unix(), TimestampBase)

	file, err := app.Update(context.Background(), 111, fid, &options)

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}

	if createdAt, exists := file.metadata[MetadataCreatedAtKey]; !exists || createdAt != createdAtValue {
		t.Errorf("got created_at = %v, want = %v", createdAt, createdAtValue)
	}

	if customField, exists := file.metadata[customFieldKey]; !exists || customField != customFieldValue {
		t.Errorf("got custom_field = %v, want = %v", customField, customFieldValue)
	}
}

func TestDeleteWhenFileDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		registerFile: func(ctx context.Context, uid int32, file *File) (string, error) {
			return "", nil
		},
	}

	fileRepo := &fileRepositoryMock{}
	app := NewFileApplication(fileRepo, dirApp, &EventBusMock{}, logger)

	userId := int32(999)
	fid := "testing"

	if _, err := app.Delete(context.Background(), userId, fid); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestDeleteWhenHasNoPermissions(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		unregisterFile: func(ctx context.Context, uid int32, file *File) error {
			return nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: make(map[int32]Permission),
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}
	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	if _, err := app.Delete(context.Background(), 999, fid); !errors.Is(err, fb.ErrNotAvailable) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotAvailable)
	}
}

func TestDeleteWhenCannotSave(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirApp := &directoryApplicationMock{
		unregisterFile: func(ctx context.Context, uid int32, file *File) error {
			return nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{222: Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},
	}

	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	if _, err := app.Delete(context.Background(), 222, fid); !errors.Is(err, fb.ErrUnknown) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrUnknown)
	}
}

func TestDeleteWhenIsNotOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryRemoveFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		unregisterFile: func(ctx context.Context, uid int32, file *File) error {
			directoryRemoveFileMethodExecuted = true
			return nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{111: Read},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},

		save: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			return nil
		},
	}

	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	file, err := app.Delete(context.Background(), 111, fid)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := file.metadata[MetadataDeletedAtKey]; exists {
		t.Errorf("got deleted_at = %v, want = %v", deletedAt, nil)
	}

	if perm, exists := file.permissions[111]; exists {
		t.Errorf("got permissions = %v, want = %v", perm, nil)
	}

	if !directoryRemoveFileMethodExecuted {
		t.Errorf("directory's RemoveFile method did not execute")
	}
}

func TestDeleteWhenMoreThanOneOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryRemoveFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		unregisterFile: func(ctx context.Context, uid int32, file *File) error {
			directoryRemoveFileMethodExecuted = true
			return nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{111: Owner, 222: Owner},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},

		save: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			return nil
		},
	}

	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	file, err := app.Delete(context.Background(), 111, fid)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := file.metadata[MetadataDeletedAtKey]; exists {
		t.Errorf("got deleted_at = %v, want = %v", deletedAt, nil)
	}

	if perm, exists := file.permissions[111]; exists {
		t.Errorf("got permissions = %v, want = %v", perm, nil)
	}

	if perm, exists := file.permissions[222]; !exists || perm != Owner {
		t.Errorf("got permissions = %v, want = %v", perm, Owner)
	}

	if !directoryRemoveFileMethodExecuted {
		t.Errorf("directory's RemoveFile method did not execute")
	}
}

func TestDeleteWhenSingleOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	directoryRemoveFileMethodExecuted := false
	dirApp := &directoryApplicationMock{
		unregisterFile: func(ctx context.Context, uid int32, file *File) error {
			directoryRemoveFileMethodExecuted = true
			return nil
		},
	}

	repo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*File, error) {
			return &File{
				id:          "123",
				name:        "testing",
				metadata:    make(Metadata),
				permissions: map[int32]Permission{111: Owner},
				data:        []byte{},
				flags:       repo.flags,
			}, nil
		},

		delete: func(repo *fileRepositoryMock, ctx context.Context, file *File) error {
			return nil
		},
	}

	app := NewFileApplication(repo, dirApp, &EventBusMock{}, logger)

	fid := "testing"
	before := time.Now().Unix()
	file, err := app.Delete(context.Background(), 111, fid)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := file.metadata[MetadataDeletedAtKey]; !exists {
		t.Errorf("got deleted_at = %v, want > %v && < %v", deletedAt, before, after)
	} else if unixDeletedAt, err := strconv.ParseInt(deletedAt, TimestampBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixDeletedAt > before || unixDeletedAt < after {
		t.Errorf("got deleted_at = %v, want > %v && < %v", unixDeletedAt, before, after)
	}

	if perm, exists := file.permissions[111]; !exists {
		t.Errorf("got permissions = %v, want = %v", perm, Owner)
	}

	if !directoryRemoveFileMethodExecuted {
		t.Errorf("directory's RemoveFile method did not execute")
	}
}
