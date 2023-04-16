package main

import (
	"net/http"
	"os"

	"github.com/alvidir/filebrowser/cmd"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"github.com/alvidir/filebrowser/user"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("loading dotenv file",
			zap.Error(err))
	}

	mongoConn := cmd.GetMongoConnection(logger)
	if header, exists := os.LookupEnv(cmd.ENV_UID_HEADER); exists {
		cmd.UidHeader = header
	}

	fileRepo := file.NewMongoFileRepository(mongoConn, logger)

	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	userApp := user.NewUserApplication(directoryRepo, fileRepo, logger)
	userService := user.NewUserRestServer(userApp, logger, cmd.UidHeader)

	lis := cmd.GetNetworkListener(logger)

	logger.Info("server ready to accept connections",
		zap.String("address", cmd.ServiceAddr))

	if err := http.Serve(lis, userService); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}
