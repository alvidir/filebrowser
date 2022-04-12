package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type FileServer struct {
	proto.UnimplementedFileServer
	app    *FileApplication
	logger *zap.Logger
	header string
}

func NewFileServer(app *FileApplication, authHeader string, logger *zap.Logger) *FileServer {
	return &FileServer{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}

func (server *FileServer) Create(ctx context.Context, req *proto.NewFile) (*proto.FileDescriptor, error) {
	ctx, err := fb.WithAuth(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Create(ctx, req.Path, req.Data, req.Metadata)
	if err != nil {
		return nil, err
	}

	return &proto.FileDescriptor{
		Id: file.id,
	}, nil
}

func (server *FileServer) Read(ctx context.Context, req *proto.FileDescriptor) (*proto.FileDescriptor, error) {
	return nil, nil
}
