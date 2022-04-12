package filebrowser

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func GetUid(ctx context.Context, header string, logger *zap.Logger) (int32, error) {
	if meta, exists := metadata.FromIncomingContext(ctx); !exists {
		return 0, ErrUnauthorized
	} else if values := meta.Get(header); len(values) == 0 {
		return 0, ErrUnauthorized
	} else if raw, err := strconv.ParseInt(values[0], 10, 32); err != nil {
		logger.Warn("parsing header data into int32",
			zap.String("header", header),
			zap.String("value", values[0]),
			zap.Error(err))

		return 0, ErrInvalidHeader
	} else {
		return int32(raw), nil
	}
}
