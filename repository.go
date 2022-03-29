package filebrowser

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// type mongoUser struct {
// 	ID     string   `bson:"_id"`
// 	UserID int32    `bson:"user_id"`
// 	Shared []string `bson:"shared,omitempty"` // paths shared with the user
// 	Files  []string `bson:"files,omitempty"`  // paths owned by the user
// }

// type mongoPath struct {
// 	ID          string          `bson:"_id"`
// 	Path        string          `bson:"path,omitempty"`
// 	Permissions map[int32]uint8 `bson:"permissions,omitempty"`
// 	Status      uint8           `bson:"status,omitempty"`
// }

// type mongoFile struct {
// 	ID       string            `bson:"_id"`
// 	PathId   string            `bson:"path_id"`
// 	Value    []byte            `bson:"value,omitempty"`
// 	Metadata map[string]string `bson:"metadata,omitempty"`
// }

type MongoRepository struct {
	Uri      string
	Database string
}

func NewMongoRepository(uri, database string) (*MongoRepository, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoRepository{uri, database}, nil
}

func (conn *MongoRepository) open(ctx context.Context) (client *mongo.Client, err error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(conn.Uri))
}
