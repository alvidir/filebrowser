package filebrowser

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

const (
	AuthKey = "auth"
)

func WithAuth(ctx context.Context, header string, logger *zap.Logger) (context.Context, error) {
	if meta, exists := metadata.FromIncomingContext(ctx); !exists {
		return nil, ErrUnauthorized
	} else if values := meta.Get(header); len(values) == 0 {
		return nil, ErrUnauthorized
	} else if raw, err := strconv.ParseInt(values[0], 10, 32); err != nil {
		logger.Warn("parsing header data into int32",
			zap.String("header", header),
			zap.String("value", values[0]),
			zap.Error(err))

		return nil, ErrInvalidHeader
	} else {
		return context.WithValue(ctx, AuthKey, int32(raw)), nil
	}
}

func GetUid(ctx context.Context, logger *zap.Logger) (int32, error) {
	uid, ok := ctx.Value(AuthKey).(int32)
	if !ok {
		logger.Error("asserting authentication id",
			zap.Any(AuthKey, ctx.Value(AuthKey)))
		return 0, ErrUnauthorized
	}

	return uid, nil
}
