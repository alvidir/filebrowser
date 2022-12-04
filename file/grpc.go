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
		Metadata:    make([]*proto.FileMetadata, 0, len(file.metadata)),
		Permissions: make([]*proto.FilePermissions, 0, len(file.permissions)),
		Flags:       uint32(file.flags),
		Data:        file.data,
	}

	for key, value := range file.metadata {
		descriptor.Metadata = append(descriptor.Metadata, &proto.FileMetadata{
			Key:   key,
			Value: value,
		})
	}

	for uid, perm := range file.permissions {
		descriptor.Permissions = append(descriptor.Permissions, &proto.FilePermissions{
			Uid:         uid,
			Permissions: NewPermissions(perm),
		})
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

func (server *FileServer) Retrieve(ctx context.Context, req *proto.FileLocator) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	file, err := server.app.Retrieve(ctx, uid, req.GetTarget())
	if err != nil {
		return nil, err
	}

	return NewFileDescriptor(file), nil
}

func (server *FileServer) Update(ctx context.Context, req *proto.FileDescriptor) (*proto.FileDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	metadata := make(Metadata)
	for _, meta := range req.GetMetadata() {
		metadata[meta.Key] = meta.Value
	}

	file, err := server.app.Update(ctx, uid, req.GetId(), req.GetName(), req.GetData(), metadata)
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

	_, err = server.app.Delete(ctx, uid, req.GetTarget())
	return nil, err
}
