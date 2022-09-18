package certificate

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	mongoCertificateCollectionName = "certificates"
)

type mongoFileAccessAuthorization struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	FileID primitive.ObjectID `bson:"file_id"`
	UserID int32              `bson:"user_id"`
	Read   bool               `bson:"can_read"`
	Write  bool               `bson:"can_write"`
	Owner  bool               `bson:"is_owner"`
	Token  []byte             `bson:"token,omitempty"`
}

func newMongoFileAccessAuthorization(cert *FileAccessCertificate) (*mongoFileAccessAuthorization, error) {
	oid := primitive.NilObjectID

	if len(cert.id) > 0 {
		var err error
		if oid, err = primitive.ObjectIDFromHex(cert.id); err != nil {
			return nil, err
		}
	}

	fid := primitive.NilObjectID
	if len(cert.fileId) > 0 {
		var err error
		if fid, err = primitive.ObjectIDFromHex(cert.fileId); err != nil {
			return nil, err
		}
	}

	return &mongoFileAccessAuthorization{
		ID:     oid,
		FileID: fid,
		UserID: cert.userId,
		Read:   cert.read,
		Write:  cert.write,
		Owner:  cert.owner,
		Token:  cert.token,
	}, nil
}

type MongoAuthorizationRepository struct {
	conn   *mongo.Collection
	logger *zap.Logger
}

func NewMongoAuthorizationRepository(db *mongo.Database, logger *zap.Logger) *MongoAuthorizationRepository {
	return &MongoAuthorizationRepository{
		conn:   db.Collection(mongoCertificateCollectionName),
		logger: logger,
	}
}

func (repo *MongoAuthorizationRepository) FindByFileIdAndUserId(ctx context.Context, fileId string, userId int32) (*FileAccessCertificate, error) {
	objID, err := primitive.ObjectIDFromHex(fileId)
	if err != nil {
		repo.logger.Error("parsing authorization id to ObjectID",
			zap.String("file_id", fileId),
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	var mAuthorization mongoFileAccessAuthorization
	err = repo.conn.FindOne(ctx, bson.M{"file_id": objID, "user_id": userId}).Decode(&mAuthorization)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("file_id", fileId),
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return repo.build(ctx, &mAuthorization)
}

func (repo *MongoAuthorizationRepository) Create(ctx context.Context, cert *FileAccessCertificate) error {
	_, err := repo.FindByFileIdAndUserId(ctx, cert.fileId, cert.userId)
	if err == nil {
		repo.logger.Warn("creating certificate",
			zap.String("file_id", cert.fileId),
			zap.Int32("user_id", cert.userId),
			zap.Error(fb.ErrAlreadyExists))

		return fb.ErrAlreadyExists
	}

	mdir, err := newMongoFileAccessAuthorization(cert)
	if err != nil {
		repo.logger.Error("building mongo authorization",
			zap.String("user_id", cert.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	res, err := repo.conn.InsertOne(ctx, mdir)
	if err != nil {
		repo.logger.Error("performing insert one on mongo",
			zap.Error(err))

		return fb.ErrUnknown
	}

	if docId, ok := res.InsertedID.(primitive.ObjectID); ok {
		cert.id = docId.Hex()
		return nil
	}

	repo.logger.Error("performing insert one on mongo",
		zap.Error(err))

	return fb.ErrUnknown
}

func (repo *MongoAuthorizationRepository) Delete(ctx context.Context, cert *FileAccessCertificate) error {
	objID, err := primitive.ObjectIDFromHex(cert.id)
	if err != nil {
		repo.logger.Error("parsing authorization id to ObjectID",
			zap.String("authorization", cert.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	result, err := repo.conn.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		repo.logger.Error("performing replace one on mongo",
			zap.Error(err))

		return fb.ErrUnknown
	}

	if result.DeletedCount == 0 {
		repo.logger.Error("performing delete one on mongo",
			zap.String("authorization_id", cert.id),
			zap.Int64("deleted_count", result.DeletedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoAuthorizationRepository) build(ctx context.Context, mdir *mongoFileAccessAuthorization) (*FileAccessCertificate, error) {
	return &FileAccessCertificate{
		id:    mdir.ID.Hex(),
		token: mdir.Token,
	}, nil
}
