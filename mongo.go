package filebrowser

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoId string

func NewMongoDBConn(uri, database string) (*mongo.Database, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return client.Database(database), nil
}
