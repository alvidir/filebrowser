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
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FileID     primitive.ObjectID `bson:"file_id"`
	UserID     int32              `bson:"user_id"`
	Permission fb.Permission      `bson:"permission"`
	Token      string             `bson:"token,omitempty"`
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
		ID:         oid,
		FileID:     fid,
		UserID:     cert.userId,
		Permission: cert.permission,
		Token:      cert.token,
	}, nil
}

type MongoCertificateRepository struct {
	conn   *mongo.Collection
	logger *zap.Logger
}

func NewMongoCertificateRepository(db *mongo.Database, logger *zap.Logger) *MongoCertificateRepository {
	return &MongoCertificateRepository{
		conn:   db.Collection(mongoCertificateCollectionName),
		logger: logger,
	}
}

func (repo *MongoCertificateRepository) FindByFileIdAndUserId(ctx context.Context, userId int32, fileId string) (*FileAccessCertificate, error) {
	objID, err := primitive.ObjectIDFromHex(fileId)
	if err != nil {
		repo.logger.Error("parsing certificate id to ObjectID",
			zap.String("file_id", fileId),
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	var mCertificate mongoFileAccessAuthorization
	err = repo.conn.FindOne(ctx, bson.M{"file_id": objID, "user_id": userId}).Decode(&mCertificate)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("file_id", fileId),
			zap.Int32("user_id", userId),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return repo.build(ctx, &mCertificate), nil
}

func (repo *MongoCertificateRepository) Create(ctx context.Context, cert *FileAccessCertificate) error {
	_, err := repo.FindByFileIdAndUserId(ctx, cert.userId, cert.fileId)
	if err == nil {
		repo.logger.Warn("creating certificate",
			zap.String("file_id", cert.fileId),
			zap.Int32("user_id", cert.userId),
			zap.Error(fb.ErrAlreadyExists))

		return fb.ErrAlreadyExists
	}

	mdir, err := newMongoFileAccessAuthorization(cert)
	if err != nil {
		repo.logger.Error("building mongo certificate",
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

func (repo *MongoCertificateRepository) Save(ctx context.Context, cert *FileAccessCertificate) error {
	mdir, err := newMongoFileAccessAuthorization(cert)
	if err != nil {
		repo.logger.Error("building mongo certificate",
			zap.String("certificate_id", cert.id),
			zap.Int32("user_id", cert.userId),
			zap.Error(err))

		return fb.ErrUnknown
	}

	if _, err = repo.conn.ReplaceOne(ctx, bson.M{"_id": mdir.ID}, mdir); err != nil {
		repo.logger.Error("performing replace one on mongo",
			zap.String("certificate_id", cert.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoCertificateRepository) Delete(ctx context.Context, cert *FileAccessCertificate) error {
	objID, err := primitive.ObjectIDFromHex(cert.id)
	if err != nil {
		repo.logger.Error("parsing certificate id to ObjectID",
			zap.String("certificate_id", cert.id),
			zap.Error(err))

		return fb.ErrUnknown
	}

	result, err := repo.conn.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		repo.logger.Error("performing delete one on mongo",
			zap.Error(err))

		return fb.ErrUnknown
	}

	if result.DeletedCount == 0 {
		repo.logger.Error("performing delete one on mongo",
			zap.String("certificate_id", cert.id),
			zap.Int64("deleted_count", result.DeletedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoCertificateRepository) build(ctx context.Context, mdir *mongoFileAccessAuthorization) *FileAccessCertificate {
	return &FileAccessCertificate{
		id:         mdir.ID.Hex(),
		fileId:     mdir.FileID.Hex(),
		userId:     mdir.UserID,
		permission: mdir.Permission,
		token:      mdir.Token,
	}
}
