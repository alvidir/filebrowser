package directory

import (
	"context"
	"errors"
	"path"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	mongoDirectoryCollectionName = "directories"
)

type mongoDirectory struct {
	ID     primitive.ObjectID            `bson:"_id,omitempty"`
	UserID int32                         `bson:"user_id"`
	Files  map[string]primitive.ObjectID `bson:"files"`
}

func newMongoDirectory(dir *Directory, logger *zap.Logger) (*mongoDirectory, error) {
	oid, err := primitive.ObjectIDFromHex(dir.id)
	if err != nil {
		logger.Error("parsing directory id to ObjectID",
			zap.String("directory", dir.id),
			zap.Int32("user", dir.userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	mongoDir := &mongoDirectory{
		ID:     oid,
		UserID: dir.userId,
		Files:  make(map[string]primitive.ObjectID),
	}

	for fpath, file := range dir.files {
		oid, err := primitive.ObjectIDFromHex(file.Id())
		if err != nil {
			logger.Error("parsing file id to ObjectID",
				zap.String("directory", dir.id),
				zap.String("file", file.Id()),
				zap.Int32("user", dir.userId),
				zap.Error(err))

			continue
		}

		mongoDir.Files[fpath] = oid
	}

	return mongoDir, nil
}

func (mdir *mongoDirectory) build() *Directory {
	dir := &Directory{
		id:     mdir.ID.Hex(),
		userId: mdir.UserID,
		files:  make(map[string]*file.File),
	}

	for fpath, oid := range mdir.Files {
		base := path.Base(fpath)
		dir.files[fpath] = file.NewFile(oid.Hex(), base, nil)
	}

	return dir
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

func (repo *MongoDirectoryRepository) FindByUserId(ctx context.Context, userId int32) (*Directory, error) {
	var mongoDirectory mongoDirectory
	err := repo.conn.FindOne(ctx, bson.M{"user_id": userId}).Decode(&mongoDirectory)
	if err != nil {
		repo.logger.Error("performing find by user id on mongo",
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return mongoDirectory.build(), nil
}

func (repo *MongoDirectoryRepository) Create(ctx context.Context, dir *Directory) error {
	var d mongoDirectory
	err := repo.conn.FindOne(ctx, bson.M{"user_id": dir.userId}).Decode(&d)
	if err == nil {
		return fb.ErrAlreadyExists
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		repo.logger.Error("performing find by user id on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	mdir, err := newMongoDirectory(dir, repo.logger)
	if err != nil {
		return err
	}

	res, err := repo.conn.InsertOne(ctx, mdir)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if docId, ok := res.InsertedID.(primitive.ObjectID); ok {
		dir.id = docId.String()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.Int32("user_id", dir.userId),
		zap.Error(err))

	return fb.ErrUnknown
}

func (repo *MongoDirectoryRepository) Save(ctx context.Context, dir *Directory) (err error) {
	mdir, err := newMongoDirectory(dir, repo.logger)
	if err != nil {
		return err
	}

	if _, err = repo.conn.ReplaceOne(ctx, bson.M{"user_id": mdir.UserID}, mdir); err != nil {
		repo.logger.Error("performing replace one on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	return
}
