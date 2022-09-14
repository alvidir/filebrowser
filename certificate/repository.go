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

type mongoCertificate struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Token string             `bson:"token"`
}

func newMongoCertificate(cert *Certificate) (*mongoCertificate, error) {
	oid := primitive.NilObjectID

	if len(cert.id) > 0 {
		var err error
		if oid, err = primitive.ObjectIDFromHex(cert.id); err != nil {
			return nil, err
		}
	}

	return &mongoCertificate{
		ID:    oid,
		Token: cert.token,
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

func (repo *MongoCertificateRepository) Find(ctx context.Context, id string) (*Certificate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		repo.logger.Error("parsing file id to ObjectID",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	var mCertificate mongoCertificate
	err = repo.conn.FindOne(ctx, bson.M{"_id": objID}).Decode(&mCertificate)
	if err != nil {
		repo.logger.Error("performing find one on mongo",
			zap.String("file_id", id),
			zap.Error(err))

		return nil, fb.ErrUnknown
	}

	return repo.build(ctx, &mCertificate)
}

func (repo *MongoCertificateRepository) Create(ctx context.Context, cert *Certificate) error {
	mdir, err := newMongoCertificate(cert)
	if err != nil {
		repo.logger.Error("building mongo certificate",
			zap.String("certificate", cert.id),
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

func (repo *MongoCertificateRepository) Delete(ctx context.Context, cert *Certificate) error {
	objID, err := primitive.ObjectIDFromHex(cert.id)
	if err != nil {
		repo.logger.Error("parsing certificate id to ObjectID",
			zap.String("certificate", cert.id),
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
			zap.String("certificate_id", cert.id),
			zap.Int64("deleted_count", result.DeletedCount))

		return fb.ErrUnknown
	}

	return nil
}

func (repo *MongoCertificateRepository) build(ctx context.Context, mdir *mongoCertificate) (*Certificate, error) {
	return &Certificate{
		id:    mdir.ID.Hex(),
		token: mdir.Token,
	}, nil
}
