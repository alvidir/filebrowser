package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type DirectoryServer struct {
	proto.UnimplementedDirectoryServer
	app    *DirectoryApplication
	logger *zap.Logger
	header string
}

func NewDirectoryServer(app *DirectoryApplication, logger *zap.Logger, authHeader string) *DirectoryServer {
	return &DirectoryServer{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}

func (server *DirectoryServer) Create(ctx context.Context, req *proto.CreateDirRequest) (*proto.CreateDirResponse, error) {
	ctx, err := fb.WithAuth(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	if _, err := server.app.Create(ctx); err != nil {
		return nil, err
	}

	return &proto.CreateDirResponse{}, nil
}
