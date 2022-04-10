package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	MongoFileCollectionName = "files"
)

type mongoFile struct {
	ID          string            `bson:"_id,omitempty"`
	Name        string            `bson:"name"`
	Flags       uint8             `bson:"flags"`
	Permissions map[int32]uint8   `bson:"permissions,omitempty"`
	Metadata    map[string]string `bson:"metadata,omitempty"`
	Data        []byte            `bson:"data,omitempty"`
}

func newMongoFile(f *File) *mongoFile {
	return &mongoFile{
		ID:          f.id,
		Name:        f.name,
		Flags:       f.flags,
		Permissions: f.permissions,
		Metadata:    f.metadata,
		Data:        f.data,
	}
}

type MongoFileRepository struct {
	conn   *mongo.Collection
	logger *zap.Logger
}

func NewMongoFileRepository(db *mongo.Database, logger *zap.Logger) *MongoFileRepository {
	return &MongoFileRepository{
		conn:   db.Collection(MongoFileCollectionName),
		logger: logger,
	}
}

func (repo *MongoFileRepository) Create(ctx context.Context, file *File) error {
	mongoFile := newMongoFile(file)
	res, err := repo.conn.InsertOne(ctx, mongoFile)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.String("filename", file.name),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if fileId, ok := res.InsertedID.(primitive.ObjectID); ok {
		file.id = fileId.Hex()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.String("filename", file.name),
		zap.Error(err))

	return fb.ErrUnknown
}
