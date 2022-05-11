package directory

import (
	"context"
	"testing"

	"github.com/alvidir/filebrowser"
	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

const (
	mockDirectoryId = "000"
)

type directoryRepositoryMock struct {
	findByUserId func(ctx context.Context, userId int32) (*Directory, error)
	create       func(ctx context.Context, dir *Directory) error
	save         func(ctx context.Context, dir *Directory) error
	delete       func(ctx context.Context, dir *Directory) error
}

func (mock *directoryRepositoryMock) FindByUserId(ctx context.Context, userId int32) (*Directory, error) {
	if mock.findByUserId != nil {
		return mock.findByUserId(ctx, userId)
	}

	dir := NewDirectory(userId)
	dir.id = mockDirectoryId
	return dir, nil
}

func (mock *directoryRepositoryMock) Create(ctx context.Context, dir *Directory) error {
	if mock.create != nil {
		return mock.create(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

func (mock *directoryRepositoryMock) Save(ctx context.Context, dir *Directory) error {
	if mock.save != nil {
		return mock.save(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

func (mock *directoryRepositoryMock) Delete(ctx context.Context, dir *Directory) error {
	if mock.delete != nil {
		return mock.delete(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

type fileRepositoryMock struct {
	create func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
	find   func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error)
	save   func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
	delete func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
	flags  uint8
}

func (mock *fileRepositoryMock) Create(ctx context.Context, file *file.File) error {
	if mock.create != nil {
		return mock.create(mock, ctx, file)
	}

	return fb.ErrNotFound
}

func (mock *fileRepositoryMock) Find(ctx context.Context, id string) (*file.File, error) {
	if mock.find != nil {
		return mock.find(mock, ctx, id)
	}

	return nil, fb.ErrNotFound
}

func (mock *fileRepositoryMock) Save(ctx context.Context, file *file.File) error {
	if mock.save != nil {
		return mock.save(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func (mock *fileRepositoryMock) Delete(ctx context.Context, file *file.File) error {
	if mock.delete != nil {
		return mock.delete(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func TestDirectoryApplication_Create(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, filebrowser.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	var want int32 = 999
	dir, err := app.Create(context.TODO(), want)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if got := dir.id; got != mockDirectoryId {
		t.Errorf("got directory.id = %v, want = %v", got, mockDirectoryId)
	}

	if got := dir.userId; got != want {
		t.Errorf("got directory.userId = %v, want = %v", got, want)
	}
}
