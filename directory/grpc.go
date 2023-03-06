package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type DirectoryServer struct {
	proto.UnimplementedDirectoryServiceServer
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

func NewProtoPath(absolute string) *proto.Path {
	if len(absolute) == 0 {
		return nil
	}

	return &proto.Path{
		Absolute: absolute,
	}
}

func NewProtoDirectory(dir *Directory) *proto.Directory {
	protoDir := &proto.Directory{
		Id:    dir.id,
		Files: make([]*proto.File, 0, len(dir.files)),
		Path:  NewProtoPath(dir.path),
	}

	for _, fs := range dir.files {
		protoDir.Files = append(protoDir.Files, file.NewProtoFile(fs))
	}

	return protoDir
}

func (server *DirectoryServer) Get(ctx context.Context, path *proto.Path) (*proto.Directory, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Get(ctx, uid, path.GetAbsolute())
	if err != nil {
		return nil, err
	}

	return NewProtoDirectory(dir), nil
}

func (server *DirectoryServer) Delete(ctx context.Context, path *proto.Path) (*proto.Directory, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Delete(ctx, uid, path.GetAbsolute())
	if err != nil {
		return nil, err
	}

	return NewProtoDirectory(dir), nil
}

func (server *DirectoryServer) Move(ctx context.Context, req *proto.MoveRequest) (*proto.Directory, error) {
	uid, err := fb.GetUid(ctx, server.header, server.logger)
	if err != nil {
		return nil, err
	}

	protoPaths := req.GetPaths()
	paths := make([]string, 0, len(protoPaths))
	for _, pp := range protoPaths {
		paths = append(paths, pp.GetAbsolute())
	}

	dir, err := server.app.Move(ctx, uid, paths, req.GetDestination().GetAbsolute())
	if err != nil {
		return nil, err
	}

	return NewProtoDirectory(dir), nil
}

func (server *Directory) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	return nil, nil
}
