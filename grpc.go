package filebrowser

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

const (
	AuthKey = "auth"
)

func WithAuth(ctx context.Context, header string) (context.Context, error) {
	if meta, exists := metadata.FromIncomingContext(ctx); !exists {
		return nil, ErrUnauthorized
	} else if values := meta.Get(header); len(values) == 0 {
		return nil, ErrUnauthorized
	} else if raw, err := strconv.ParseInt(values[0], 10, 32); err != nil {
		return nil, ErrInvalidHeader
	} else {
		return context.WithValue(ctx, AuthKey, int32(raw)), nil
	}
}
