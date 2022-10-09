package certificate

import (
	"context"

	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

type CertificateRepository interface {
	Create(ctx context.Context, cert *FileAccessCertificate) error
	FindByFileIdAndUserId(ctx context.Context, userId int32, fileId string) (*FileAccessCertificate, error)
	Save(ctx context.Context, cert *FileAccessCertificate) error
	Delete(ctx context.Context, cert *FileAccessCertificate) error
}

type CertificateService interface {
	SignFileAccessCertificate(cert *FileAccessCertificate) error
}

type CertificateApplication struct {
	certRepo    CertificateRepository
	certService CertificateService
	logger      *zap.Logger
}

func NewCertificateApplication(repo CertificateRepository, srv CertificateService, logger *zap.Logger) *CertificateApplication {
	return &CertificateApplication{
		certRepo:    repo,
		certService: srv,
		logger:      logger,
	}
}

func (app *CertificateApplication) CreateFileAccessCertificate(ctx context.Context, uid int32, file *file.File) (*FileAccessCertificate, error) {
	cert := NewFileAccessCertificate(uid, file)
	if err := app.certRepo.Create(ctx, cert); err != nil {
		return nil, err
	}

	if err := app.certService.SignFileAccessCertificate(cert); err != nil {
		return nil, err
	}

	if err := app.certRepo.Save(ctx, cert); err != nil {
		return nil, err
	}

	return cert, nil
}

func (app *CertificateApplication) GetFileAccessCertificate(ctx context.Context, uid int32, fid string) (*FileAccessCertificate, error) {
	return app.certRepo.FindByFileIdAndUserId(ctx, uid, fid)
}
