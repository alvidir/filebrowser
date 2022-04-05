package file

import (
	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MongoFileCollectionName = "files"
)

type MongoFile struct {
	ID          fb.MongoId           `bson:"_id"`
	Name        string               `bson:"name"`
	Flags       uint8                `bson:"flags"`
	Permissions map[fb.MongoId]uint8 `bson:"permissions"`
	Metadata    map[string]string    `bson:"metadata,omitempty"`
	Value       []byte               `bson:"value,omitempty"`
}

type MongoFileRepository struct {
	conn *mongo.Collection
}

func NewMongoFileRepository(db *mongo.Database) *MongoFileRepository {
	return &MongoFileRepository{
		conn: db.Collection(MongoFileCollectionName),
	}
}
