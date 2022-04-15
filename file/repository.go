package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	MongoFileCollectionName = "files"
)

type mongoFile struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Flags       uint8              `bson:"flags"`
	Permissions map[int32]uint8    `bson:"permissions,omitempty"`
	Metadata    map[string]string  `bson:"metadata,omitempty"`
	Data        []byte             `bson:"data,omitempty"`
}

func newMongoFile(f *File, logger *zap.Logger) (*mongoFile, error) {
	oid, err := primitive.ObjectIDFromHex(f.id)
	if err != nil {
		logger.Error("parsing file id to ObjectID",
			zap.String("directory", f.id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return &mongoFile{
		ID:          oid,
		Name:        f.name,
		Flags:       f.flags,
		Permissions: f.permissions,
		Metadata:    f.metadata,
		Data:        f.data,
	}, nil
}

func (mfile *mongoFile) build() *File {
	return &File{
		id:          mfile.ID.Hex(),
		name:        mfile.Name,
		metadata:    mfile.Metadata,
		permissions: mfile.Permissions,
		flags:       mfile.Flags,
		data:        mfile.Data,
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
	mongoFile, err := newMongoFile(file, repo.logger)
	if err != nil {
		return err
	}

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

func (repo *MongoFileRepository) Find(ctx context.Context, id string) (*File, error) {
	objID, _ := primitive.ObjectIDFromHex(id)

	var mfile mongoFile
	err := repo.conn.FindOne(ctx, bson.M{"_id": objID}).Decode(&mfile)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("fileid", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return mfile.build(), nil
}
