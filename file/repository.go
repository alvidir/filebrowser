package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func newMongoFile(f *File) (*mongoFile, error) {
	oid := primitive.NilObjectID

	if len(f.id) > 0 {
		var err error
		if oid, err = primitive.ObjectIDFromHex(f.id); err != nil {
			return nil, err
		}
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
	mongoFile, err := newMongoFile(file)
	if err != nil {
		repo.logger.Error("building mongo file",
			zap.String("file_id", file.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	res, err := repo.conn.InsertOne(ctx, mongoFile)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.String("file_name", file.name),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if fileId, ok := res.InsertedID.(primitive.ObjectID); ok {
		file.id = fileId.Hex()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.String("file_name", file.name),
		zap.Error(err))

	return fb.ErrUnknown
}

func (repo *MongoFileRepository) Find(ctx context.Context, id string) (*File, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.logger.Error("parsing file id to ObjectID",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	var mfile mongoFile
	err = repo.conn.FindOne(ctx, bson.M{"_id": objID}).Decode(&mfile)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return repo.build(&mfile), nil
}

// FindAll returns all those files matching the given ids, excluding the data field
func (repo *MongoFileRepository) FindAll(ctx context.Context, ids []string) ([]*File, error) {
	objIDs := make([]primitive.ObjectID, len(ids))
	for index, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			repo.logger.Error("parsing file id to ObjectID",
				zap.String("file_id", id),
				zap.Error(err))

			return nil, fb.ErrUnknown
		}

		objIDs[index] = objID
	}

	// exclude data field from being loaded
	opts := options.Find().SetProjection(bson.D{{Key: "data", Value: 0}})
	cursor, err := repo.conn.Find(ctx, bson.M{"_id": bson.M{"$in": objIDs}}, opts)
	if err != nil {
		repo.logger.Error("performing find all on mongo",
			zap.Strings("file_ids", ids),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	mfiles := make([]mongoFile, len(ids))
	if err := cursor.All(ctx, &mfiles); err != nil {
		repo.logger.Error("decoding found items",
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	files := make([]*File, len(ids))
	for index, mfile := range mfiles {
		files[index] = repo.build(&mfile)
	}

	return files, nil
}

func (repo *MongoFileRepository) FindPermissions(ctx context.Context, id string) (Permissions, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.logger.Error("parsing file id to ObjectID",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	// exclude all fields except permissions
	opts := options.FindOne().SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "name", Value: 0},
		{Key: "flags", Value: 0},
		{Key: "metadata", Value: 0},
		{Key: "data", Value: 0},
	})

	var mfile mongoFile
	err = repo.conn.FindOne(ctx, bson.M{"_id": objID}, opts).Decode(&mfile)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return mfile.Permissions, nil
}

func (repo *MongoFileRepository) Save(ctx context.Context, file *File) error {
	mFile, err := newMongoFile(file)
	if err != nil {
		repo.logger.Error("building mongo file",
			zap.String("file_id", file.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	result, err := repo.conn.ReplaceOne(ctx, bson.M{"_id": mFile.ID}, mFile)
	if err != nil {
		repo.logger.Error("performing replace one on mongo",
			zap.String("file_id", file.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if result.ModifiedCount == 0 {
		repo.logger.Error("performing replace one on mongo",
			zap.String("file_id", file.id),
			zap.Int64("modified_count", result.ModifiedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoFileRepository) Delete(ctx context.Context, file *File) error {
	objID, err := primitive.ObjectIDFromHex(file.id)
	if err != nil {
		repo.logger.Error("parsing file id to ObjectID",
			zap.String("file_id", file.Id()),
			zap.Error(err))

		return fb.ErrUnknown
	}

	result, err := repo.conn.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		repo.logger.Error("performing delete one on mongo",
			zap.String("file_id", file.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if result.DeletedCount == 0 {
		repo.logger.Error("performing delete one on mongo",
			zap.String("file_id", file.id),
			zap.Int64("deleted_count", result.DeletedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoFileRepository) build(mfile *mongoFile) *File {
	return &File{
		id:          mfile.ID.Hex(),
		name:        mfile.Name,
		metadata:    mfile.Metadata,
		permissions: mfile.Permissions,
		flags:       mfile.Flags,
		data:        mfile.Data,
	}
}
