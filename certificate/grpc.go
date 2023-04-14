package certificate

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type CertificateGrpcService struct {
	proto.UnimplementedCertificateServiceServer
	app    *CertificateApplication
	logger *zap.Logger
	header string
}

func NewCertificateGrpcServer(app *CertificateApplication, logger *zap.Logger, authHeader string) *CertificateGrpcService {
	return &CertificateGrpcService{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}

func (server *CertificateGrpcService) Get(ctx context.Context, req *proto.File) (*proto.Certificate, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	cert, err := server.app.GetFileAccessCertificate(ctx, uid, req.GetId())
	if err != nil {
		return nil, err
	}

	descriptor := &proto.Certificate{
		Id: cert.id,
		Permissions: &proto.Permissions{
			Read:  cert.permission&Read != 0,
			Write: cert.permission&Write != 0,
			Owner: cert.permission&Owner != 0,
		},
		Token: cert.token,
	}

	return descriptor, nil
}
