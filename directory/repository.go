package directory

import (
	"context"
	"errors"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
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
	conn   *mongo.Collection
	logger *zap.Logger
}

func NewMongoDirectoryRepository(db *mongo.Database, logger *zap.Logger) *MongoDirectoryRepository {
	return &MongoDirectoryRepository{
		conn:   db.Collection(mongoDirectoryCollectionName),
		logger: logger,
	}
}

func (repo *MongoDirectoryRepository) Create(ctx context.Context, dir *Directory) error {
	var mongoDirectory MongoDirectory
	err := repo.conn.FindOne(ctx, bson.M{"user_id": dir.userId}).Decode(&mongoDirectory)
	if err == nil {
		return fb.ErrAlreadyExists
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		repo.logger.Error("performing find by user id on mongo",
			zap.Int32("userId", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	mongoDirectory.UserID = dir.userId
	res, err := repo.conn.InsertOne(ctx, mongoDirectory)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.Int32("userId", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if docId, ok := res.InsertedID.(primitive.ObjectID); ok {
		dir.id = docId.String()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.Int32("userId", dir.userId),
		zap.Error(err))

	return fb.ErrUnknown
}
