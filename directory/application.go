package directory

import "context"

type DirectoryRepository interface {
	FindByUserId(ctx context.Context, userID int32) (*Directory, error)
}

type DirectoryApplication struct {
	directoryRepo DirectoryRepository
}

func (app *DirectoryApplication) Create(ctx context.Context, userId int32) error {
	return nil
}

func (app *DirectoryApplication) Delete(ctx context.Context, id string) error {
	return nil
}
