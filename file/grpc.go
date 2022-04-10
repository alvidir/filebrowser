package file

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type FileServer struct {
	proto.UnimplementedFileServer
	fileApp    *FileApplication
	logger     *zap.Logger
	authHeader string
}

func NewFileServer(app *FileApplication, logger *zap.Logger, authHeader string) *FileServer {
	return &FileServer{
		fileApp:    app,
		logger:     logger,
		authHeader: authHeader,
	}
}

func (server *FileServer) Create(ctx context.Context, req *proto.CreateFileRequest) (*proto.CreateFileResponse, error) {
	ctx, err := fb.WithAuth(ctx, server.authHeader)
	if err != nil {
		server.logger.Warn("getting users authentication",
			zap.String("header", server.authHeader),
			zap.Error(err))
		return nil, err
	}

	file, err := server.fileApp.Create(ctx, req.Path, req.Data, req.Metadata)
	if err != nil {
		return nil, err
	}

	return &proto.CreateFileResponse{
		Id: file.id,
	}, nil
}
