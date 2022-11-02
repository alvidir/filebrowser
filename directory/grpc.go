package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
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

func (server *DirectoryServer) Create(ctx context.Context, req *proto.DirectoryLocator) (*proto.DirectoryDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Create(ctx, uid)
	if err != nil {
		return nil, err
	}

	descriptor := &proto.DirectoryDescriptor{
		Id:    dir.id,
		Files: make([]*proto.FileDescriptor, 0, len(dir.Files())),
	}

	for _, fs := range dir.Files() {
		descriptor.Files = append(descriptor.Files, &proto.FileDescriptor{
			Id:       fs.Id(),
			Metadata: fs.Metadata(),
		})
	}

	return descriptor, nil
}

func (server *DirectoryServer) Retrieve(ctx context.Context, req *proto.DirectoryLocator) (*proto.DirectoryDescriptor, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Retrieve(ctx, uid, req.GetPath())
	if err != nil {
		return nil, err
	}

	descriptor := &proto.DirectoryDescriptor{
		Id:    dir.id,
		Files: make([]*proto.FileDescriptor, 0, len(dir.Files())),
	}

	for _, fs := range dir.Files() {
		descriptor.Files = append(descriptor.Files, file.NewFileDescriptor(fs))
	}

	return descriptor, nil
}
