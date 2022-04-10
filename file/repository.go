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

type MongoFile struct {
	ID          string            `bson:"_id"`
	Name        string            `bson:"name"`
	Flags       uint8             `bson:"flags"`
	Permissions map[string]uint8  `bson:"permissions"`
	Metadata    map[string]string `bson:"metadata,omitempty"`
	Value       []byte            `bson:"value,omitempty"`
}

func newMongoFile(f *File) *MongoFile {
	return &MongoFile{
		ID:          f.id,
		Name:        f.name,
		Flags:       f.flags,
		Permissions: f.permissions,
		Metadata:    f.metadata,
		Value:       f.value,
	}
}

type MongoFileRepository struct {
	conn   *mongo.Collection
	logger *zap.Logger
}

func NewMongoFileRepository(db *mongo.Database) *MongoFileRepository {
	return &MongoFileRepository{
		conn: db.Collection(MongoFileCollectionName),
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
		file.id = fileId.String()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.String("filename", file.name),
		zap.Error(err))

	return fb.ErrUnknown
}
