package file

import (
	"context"
)

const (
	mockDirectoryId = "000"
)

type fileRepositoryMock struct {
	create func(ctx context.Context, dir *File) error
}

func (mock *fileRepositoryMock) Create(ctx context.Context, file *File) error {
	if mock.create != nil {
		return mock.create(ctx, file)
	}

	file.id = mockDirectoryId
	return nil
}

// func TestFileApplication_create(t *testing.T) {
// 	logger, _ := zap.NewProduction()
// 	defer logger.Sync()

// 	repo := &fileRepositoryMock{}
// 	app := NewFileApplication(repo, logger)

// 	var userId int32 = 999
// 	ctx := context.WithValue(context.TODO(), fb.AuthKey, userId)
// }
