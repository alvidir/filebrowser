package directory

import (
	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUser struct {
	ID     fb.MongoId `bson:"_id"`
	UserID int32      `bson:"user_id"`
}

type MongoDirectory struct {
	ID     fb.MongoId            `bson:"_id"`
	Shared map[string]fb.MongoId `bson:"shared"`
	Hosted map[string]fb.MongoId `bson:"hosted"`
}

type MongoDirectoryRepository struct {
	conn *mongo.Database
}
