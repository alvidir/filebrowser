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
	dirBus chan<- interface{}
	header string
}

func NewFileServer(app *FileApplication, authHeader string, logger *zap.Logger) *FileServer {
	return &FileServer{
		app:    app,
		logger: logger,
		dirBus: nil,
		header: authHeader,
	}
}

func (server *FileServer) sendFileCreatedEvent(ctx context.Context, fileId, path string) {
	if server.dirBus == nil {
		return
	}

	ctx = context.WithValue(context.Background(), fb.AuthKey, ctx.Value(fb.AuthKey))
	server.dirBus <- newFileCreatedEvent(ctx, fileId, path)
}

func (server *FileServer) RegisterDirectoryEventBus(out chan<- interface{}) {
	server.dirBus = out
}

func (server *FileServer) Create(ctx context.Context, req *proto.CreateFileRequest) (*proto.CreateFileResponse, error) {
	ctx, err := fb.WithAuth(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Create(ctx, req.Path, req.Data, req.Metadata)
	if err != nil {
		return nil, err
	}

	server.sendFileCreatedEvent(ctx, file.id, req.Path)
	return &proto.CreateFileResponse{
		Id: file.id,
	}, nil
}
