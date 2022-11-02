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

func NewFileDescriptor(file *File) *proto.FileDescriptor {
	descriptor := &proto.FileDescriptor{
		Id:          file.id,
		Name:        file.name,
		Metadata:    file.metadata,
		Permissions: map[int32]*proto.Permissions{},
		Flags:       uint32(file.flags),
		Data:        file.data,
	}

	for uid, perm := range file.permissions {
		descriptor.Permissions[uid] = NewPermissions(perm)
	}

	return descriptor
}

func NewPermissions(perm fb.Permission) *proto.Permissions {
	return &proto.Permissions{
		Read:  perm&fb.Read != 0,
		Write: perm&fb.Write != 0,
		Owner: perm&fb.Owner != 0,
	}
}

func NewFileServer(app *FileApplication, authHeader string, logger *zap.Logger) *FileServer {
	return &FileServer{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}

func (server *FileServer) Create(ctx context.Context, req *proto.FileConstructor) (*proto.FileDescriptor, error) {
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

func (server *FileServer) Read(ctx context.Context, req *proto.FileLocator) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Read(ctx, uid, req.GetId())
	if err != nil {
		return nil, err
	}

	return NewFileDescriptor(file), nil
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

	return NewFileDescriptor(file), nil
}

func (server *FileServer) Delete(ctx context.Context, req *proto.FileLocator) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	_, err = server.app.Delete(ctx, uid, req.GetId())
	return nil, err
}
