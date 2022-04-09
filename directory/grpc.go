package directory

import (
	"context"

	proto "github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type DirectoryServer struct {
	proto.UnimplementedDirectoryServer
	directoryApp *DirectoryApplication
	logger       *zap.Logger
}

func NewDirectoryServer(app *DirectoryApplication, logger *zap.Logger) *DirectoryServer {
	return &DirectoryServer{
		directoryApp: app,
		logger:       logger,
	}
}

func (server *DirectoryServer) Create(context.Context, *proto.Empty) (*proto.Empty, error) {
	return nil, nil
}
