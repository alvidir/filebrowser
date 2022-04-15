package file

import (
	"context"
	"path"
	"strconv"
	"testing"
	"time"

	"go.uber.org/zap"
)

const (
	mockFileId = "000"
)

type fileEventHandlerMock struct {
	onFileCreated func(file *File, uid int32, path string)
	ch            chan struct {
		file *File
		uid  int32
		path string
	}

	t *testing.T
}

func newFileEventHandlerMock(t *testing.T) *fileEventHandlerMock {
	return &fileEventHandlerMock{
		ch: make(chan struct {
			file *File
			uid  int32
			path string
		}),

		t: t,
	}
}

func (handler *fileEventHandlerMock) OnFileCreated(file *File, uid int32, path string) {
	if handler.onFileCreated != nil {
		handler.onFileCreated(file, uid, path)
		return
	}

	handler.ch <- struct {
		file *File
		uid  int32
		path string
	}{file, uid, path}
}

type fileRepositoryMock struct {
	create func(ctx context.Context, dir *File) error
	find   func(ctx context.Context, id string) (*File, error)
}

func (mock *fileRepositoryMock) Create(ctx context.Context, file *File) error {
	if mock.create != nil {
		return mock.create(ctx, file)
	}

	file.id = mockFileId
	return nil
}

func (mock *fileRepositoryMock) Find(ctx context.Context, id string) (*File, error) {
	return nil, nil
}

func TestFileApplication_create(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := &fileRepositoryMock{}
	handler := newFileEventHandlerMock(t)

	app := NewFileApplication(repo, logger)
	app.Subscribe(handler)

	userId := int32(999)
	fpath := "path/to/example.test"
	data := []byte{1, 2, 3, 4}
	meta := make(Metadata)

	customFieldKey := "custom_field"
	customFieldValue := "custom value"
	meta[customFieldKey] = customFieldValue

	before := time.Now().Unix()
	file, err := app.Create(context.Background(), userId, fpath, data, meta)
	after := time.Now().Unix()

	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
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

	ticker := time.NewTicker(10 * time.Second)
	select {
	case v := <-handler.ch:
		if v.file != file {
			t.Errorf("got field = %p, want = %p", v.file, file)
		}

		if v.uid != userId {
			t.Errorf("got user id = %v, want = %v", v.uid, userId)
		}

		if v.path != fpath {
			t.Errorf("got path = %v, want = %v", v.path, fpath)
		}
	case <-ticker.C:
		t.Errorf("timeout exceed")
	}
}
