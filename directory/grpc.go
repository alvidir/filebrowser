package directory

import (
	"context"
	"strconv"

	fb "github.com/alvidir/filebrowser"
	proto "github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type DirectoryServer struct {
	proto.UnimplementedDirectoryServer
	directoryApp *DirectoryApplication
	logger       *zap.Logger
	authHeader   string
}

func NewDirectoryServer(app *DirectoryApplication, logger *zap.Logger, authHeader string) *DirectoryServer {
	return &DirectoryServer{
		directoryApp: app,
		logger:       logger,
		authHeader:   authHeader,
	}
}

func (server *DirectoryServer) Create(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	var userId int32
	if meta, exists := metadata.FromIncomingContext(ctx); !exists {
		return nil, fb.ErrUnauthorized
	} else if values := meta.Get(server.authHeader); len(values) == 0 {
		return nil, fb.ErrUnauthorized
	} else if raw, err := strconv.ParseInt(values[0], 10, 32); err != nil {
		server.logger.Warn("parsing header into int32",
			zap.String("header", server.authHeader),
			zap.String("value", values[0]))
		return nil, fb.ErrInvalidHeader
	} else {
		userId = int32(raw)
	}

	if _, err := server.directoryApp.Create(ctx, userId); err != nil {
		return nil, err
	}

	return &proto.Empty{}, nil
}
