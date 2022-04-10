package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type DirectoryServer struct {
	proto.UnimplementedDirectoryServer
	directoryApp *DirectoryApplication
	logger       *zap.Logger
	authHeader   string
}

func NewDirectoryServer(app *DirectoryApplication, logger *zap.Logger, authHeader string) *DirectoryServer {
	return &DirectoryServer{
		directoryApp: app,
		logger:       logger,
		authHeader:   authHeader,
	}
}

func (server *DirectoryServer) Create(ctx context.Context, req *proto.CreateDirRequest) (*proto.CreateDirResponse, error) {
	ctx, err := fb.WithAuth(ctx, server.authHeader)
	if err != nil {
		server.logger.Warn("getting users authentication",
			zap.String("header", server.authHeader),
			zap.Error(err))
		return nil, err
	}

	if _, err := server.directoryApp.Create(ctx); err != nil {
		return nil, err
	}

	return &proto.CreateDirResponse{}, nil
}
