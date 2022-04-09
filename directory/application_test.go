package directory

import (
	"context"
	"testing"

	"github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

const (
	mockDirectoryId = "000"
)

type directoryRepositoryMock struct {
	findByUserId func(ctx context.Context, userId int32) (*Directory, error)
	create       func(ctx context.Context, dir *Directory) error
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

func TestDirectoryApplication_create(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	repo := &directoryRepositoryMock{}
	repo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, filebrowser.ErrNotFound
	}

	app := NewDirectoryApplication(repo, logger)

	var want int32 = 999
	dir, err := app.Create(context.Background(), want)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if got := dir.userId; got != want {
		t.Errorf("got directory.userId = %v, want = %v", got, want)
	}

	if got := dir.id; got != mockDirectoryId {
		t.Errorf("got directory.id = %v, want = %v", got, mockDirectoryId)
	}
}
