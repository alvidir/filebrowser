package file

import "context"

type FileRepository interface {
	FindByUserId(ctx context.Context, userID int32) (*File, error)
}

type FileApplication struct {
	FileRepo FileRepository
}

func (app *FileApplication) Create(ctx context.Context, userId int32) error {
	return nil
}
