package directory

import (
	"context"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type DirectoryGrpcServer struct {
	proto.UnimplementedDirectoryServiceServer
	app       *DirectoryApplication
	logger    *zap.Logger
	uidHeader string
}

func NewDirectoryGrpcServer(app *DirectoryApplication, logger *zap.Logger, authHeader string) *DirectoryGrpcServer {
	return &DirectoryGrpcServer{
		app:       app,
		logger:    logger,
		uidHeader: authHeader,
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

func NewProtoSearchMatch(match SearchMatch) *proto.SearchMatch {
	return &proto.SearchMatch{
		File:       file.NewProtoFile(match.file),
		MatchStart: int32(match.start),
		MatchEnd:   int32(match.end),
	}
}

func NewProtoSearchResponse(matches []SearchMatch) *proto.SearchResponse {
	search := &proto.SearchResponse{
		Matches: make([]*proto.SearchMatch, 0, len(matches)),
	}

	for _, match := range matches {
		match := NewProtoSearchMatch(match)
		search.Matches = append(search.Matches, match)
	}

	return search
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

func (server *DirectoryGrpcServer) Get(ctx context.Context, path *proto.Path) (*proto.Directory, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Get(ctx, uid, path.GetAbsolute())
	if err != nil {
		return nil, err
	}

	return NewProtoDirectory(dir), nil
}

func (server *DirectoryGrpcServer) Delete(ctx context.Context, path *proto.Path) (*proto.Directory, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	dir, err := server.app.Delete(ctx, uid, path.GetAbsolute())
	if err != nil {
		return nil, err
	}

	return NewProtoDirectory(dir), nil
}

func (server *DirectoryGrpcServer) Move(ctx context.Context, req *proto.MoveRequest) (*proto.Directory, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
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

func (server *DirectoryGrpcServer) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	uid, err := fb.GetUidFromGrpcCtx(ctx, server.uidHeader, server.logger)
	if err != nil {
		return nil, err
	}

	search, err := server.app.Search(ctx, uid, req.GetSearch())
	if err != nil {
		return nil, err
	}

	return NewProtoSearchResponse(search), nil
}
