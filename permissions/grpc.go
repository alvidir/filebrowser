package permissions

import (
	"github.com/alvidir/filebrowser/proto"
	"go.uber.org/zap"
)

type PermissionsServer struct {
	proto.UnimplementedPermissionsServer
	app    *PermissionsApplication
	logger *zap.Logger
	header string
}

func NewPermissionsServer(app *PermissionsApplication, logger *zap.Logger, authHeader string) *PermissionsServer {
	return &PermissionsServer{
		app:    app,
		logger: logger,
		header: authHeader,
	}
}
