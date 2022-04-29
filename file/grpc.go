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
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Create(ctx, uid, req.GetPath(), req.GetData(), req.GetMetadata())
	if err != nil {
		return nil, err
	}

	return &proto.FileDescriptor{
		Id: file.id,
	}, nil
}

func (server *FileServer) Read(ctx context.Context, req *proto.FileDescriptor) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Read(ctx, uid, req.GetId())
	if err != nil {
		return nil, err
	}

	descriptor := &proto.FileDescriptor{
		Id:          file.id,
		Name:        file.name,
		Metadata:    file.metadata,
		Permissions: make(map[int32]int32),
		Data:        file.data,
	}

	for uid, perm := range file.permissions {
		descriptor.Permissions[uid] = int32(perm)
	}

	return descriptor, nil
}

func (server *FileServer) Write(ctx context.Context, req *proto.FileDescriptor) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Write(ctx, uid, req.GetId(), req.GetData(), req.GetMetadata())
	if err != nil {
		return nil, err
	}

	descriptor := &proto.FileDescriptor{
		Id:          file.id,
		Name:        file.name,
		Metadata:    file.metadata,
		Permissions: make(map[int32]int32),
		Data:        file.data,
	}

	for uid, perm := range file.permissions {
		descriptor.Permissions[uid] = int32(perm)
	}

	return descriptor, nil
}
