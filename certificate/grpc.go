package certificate

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type CertificateServer struct {
	proto.UnimplementedCertificateServer
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

func (server *CertificateServer) Retrieve(ctx context.Context, req *proto.CertificateLocator) (*proto.CertificateDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	cert, err := server.app.GetFileAccessCertificate(ctx, uid, req.FileId)
	if err != nil {
		return nil, err
	}

	descriptor := &proto.CertificateDescriptor{
		Id: cert.id,
		Permissions: &proto.FilePermissions{
			Read:  cert.permissions&fb.Read != 0,
			Write: cert.permissions&fb.Write != 0,
			Owner: cert.permissions&fb.Owner != 0,
		},
		Token: cert.token,
	}

	return descriptor, nil
}
