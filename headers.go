package filebrowser

import (
	"context"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func intoInt32(s string) (int32, error) {
	if raw, err := strconv.ParseInt(s, 10, 32); err != nil {
		return 0, ErrInvalidHeader
	} else {
		return int32(raw), nil
	}
}

func GetUidFromGrpcCtx(ctx context.Context, header string, logger *zap.Logger) (int32, error) {
	if meta, exists := metadata.FromIncomingContext(ctx); !exists {
		return 0, ErrUnauthorized
	} else if values := meta.Get(header); len(values) == 0 {
		return 0, ErrUnauthorized
	} else if uid, err := intoInt32(values[0]); err != nil {
		logger.Warn("parsing header data into int32",
			zap.String("header", header),
			zap.String("value", values[0]),
			zap.Error(err))

		return 0, ErrInvalidHeader
	} else {
		return uid, nil
	}
}

func GetUidFromHttpRequest(r *http.Request, header string, logger *zap.Logger) (int32, error) {
	if values, exits := r.Header[header]; !exits || len(values) == 0 {
		return 0, ErrUnauthorized
	} else if uid, err := intoInt32(values[0]); err != nil {
		logger.Warn("parsing header data into int32",
			zap.String("header", header),
			zap.String("value", values[0]),
			zap.Error(err))

		return 0, ErrInvalidHeader
	} else {
		return uid, nil
	}
}
