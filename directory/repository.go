package directory

import (
	"context"
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

func newMongoDirectory(dir *Directory) (*mongoDirectory, error) {
	oid := primitive.NilObjectID

	if len(dir.id) > 0 {
		var err error
		if oid, err = primitive.ObjectIDFromHex(dir.id); err != nil {
			return nil, err
		}
	}

	mongoDir := &mongoDirectory{
		ID:     oid,
		UserID: dir.userId,
		Files:  make(map[string]primitive.ObjectID),
	}

	for fpath, f := range dir.files {
		oid, err := primitive.ObjectIDFromHex(f.Id())
		if err != nil {
			return nil, err
		}

		mongoDir.Files[fpath] = oid
	}

	return mongoDir, nil
}

type MongoDirectoryRepository struct {
	fileRepo file.FileRepository
	conn     *mongo.Collection
	logger   *zap.Logger
}

func NewMongoDirectoryRepository(db *mongo.Database, fileRepo file.FileRepository, logger *zap.Logger) *MongoDirectoryRepository {
	return &MongoDirectoryRepository{
		fileRepo: fileRepo,
		conn:     db.Collection(mongoDirectoryCollectionName),
		logger:   logger,
	}
}

func (repo *MongoDirectoryRepository) FindByUserId(ctx context.Context, userId int32, options *RepoOptions) (*Directory, error) {
	var mdir mongoDirectory
	err := repo.conn.FindOne(ctx, bson.M{"user_id": userId}).Decode(&mdir)
	if err != nil {
		repo.logger.Error("performing find by user id on mongo",
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return repo.build(ctx, &mdir, options)
}

func (repo *MongoDirectoryRepository) Create(ctx context.Context, dir *Directory) error {
	mdir, err := newMongoDirectory(dir)
	if err != nil {
		repo.logger.Error("building mongo directory",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	res, err := repo.conn.InsertOne(ctx, mdir)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if docId, ok := res.InsertedID.(primitive.ObjectID); ok {
		dir.id = docId.Hex()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.Int32("user_id", dir.userId),
		zap.Error(err))

	return fb.ErrUnknown
}

func (repo *MongoDirectoryRepository) Save(ctx context.Context, dir *Directory) error {
	mdir, err := newMongoDirectory(dir)
	if err != nil {
		repo.logger.Error("building mongo directory",
			zap.String("directory_id", dir.id),
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if _, err = repo.conn.ReplaceOne(ctx, bson.M{"_id": mdir.ID}, mdir); err != nil {
		repo.logger.Error("performing replace one on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoDirectoryRepository) Delete(ctx context.Context, dir *Directory) error {
	objID, err := primitive.ObjectIDFromHex(dir.id)
	if err != nil {
		repo.logger.Error("parsing directory id to ObjectID",
			zap.String("directory_id", dir.id),
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	result, err := repo.conn.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		repo.logger.Error("performing delete one on mongo",
			zap.Int32("user_id", dir.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if result.DeletedCount == 0 {
		repo.logger.Error("performing delete one on mongo",
			zap.String("directory_id", dir.id),
			zap.Int64("deleted_count", result.DeletedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoDirectoryRepository) build(ctx context.Context, mdir *mongoDirectory, options *RepoOptions) (*Directory, error) {
	dir := &Directory{
		id:     mdir.ID.Hex(),
		userId: mdir.UserID,
		files:  make(map[string]*file.File),
	}

	if options == nil || options.LazyLoading {
		for fpath, oid := range mdir.Files {
			f, _ := file.NewFile(oid.Hex(), path.Base(fpath))
			dir.files[fpath] = f
		}
	} else {

		filesIds := make([]string, 0, len(mdir.Files))
		pathByFileId := make(map[string]string)
		for fpath, oid := range mdir.Files {
			filesIds = append(filesIds, oid.Hex())
			pathByFileId[oid.Hex()] = fpath
		}

		files, err := repo.fileRepo.FindAll(ctx, filesIds)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f == nil {
				continue
			}

			dir.files[pathByFileId[f.Id()]] = f
		}
	}

	return dir, nil
}
