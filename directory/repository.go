package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mongoDirectoryCollectionName = "directories"
)

type MongoDirectory struct {
	ID     fb.MongoId            `bson:"_id;omitempty"`
	UserID int32                 `bson:"user_id"`
	Shared map[string]fb.MongoId `bson:"shared"`
	Hosted map[string]fb.MongoId `bson:"hosted"`
}

type MongoDirectoryRepository struct {
	conn *mongo.Collection
}

func NewMongoDirectoryRepository(db *mongo.Database) *MongoDirectoryRepository {
	return &MongoDirectoryRepository{
		conn: db.Collection(mongoDirectoryCollectionName),
	}
}

func (repo *MongoDirectoryRepository) FindByUserId(ctx context.Context, userId int32) (*Directory, error) {
	return nil, nil
}

func (repo *MongoDirectoryRepository) Create(ctx context.Context, directory *Directory) error {
	return nil
}
