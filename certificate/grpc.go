package certificate

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type CertificateServer struct {
	proto.UnimplementedCertificateServiceServer
	app    *CertificateApplication
	logger *zap.Logger
	header string
}

func NewCertificateServer(app *CertificateApplication, logger *zap.Logger, authHeader string) *CertificateServer {
	return &CertificateServer{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}

func (server *CertificateServer) Get(ctx context.Context, req *proto.File) (*proto.Certificate, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
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
