package main

import (
	"net"
	"os"

	fb "github.com/alvidir/filebrowser"
	dir "github.com/alvidir/filebrowser/directory"
	file "github.com/alvidir/filebrowser/file"
	proto "github.com/alvidir/filebrowser/proto"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	ENV_SERVICE_PORT   = "SERVICE_PORT"
	ENV_SERVICE_NETW   = "SERVICE_NETW"
	ENV_AUTH_HEADER    = "AUTH_HEADER"
	ENV_MONGO_DSN      = "MONGO_DSN"
	ENV_MONGO_DATABASE = "MONGO_INITDB_DATABASE"
)

var (
	servicePort = "8000"
	serviceNetw = "tcp"
	authHeader  = "X-Auth"
)

func startServer(lis net.Listener, logger *zap.Logger) {
	mongoUri, exists := os.LookupEnv(ENV_MONGO_DSN)
	if !exists {
		logger.Fatal("mongo dsn must be set")
	}

	database, exists := os.LookupEnv(ENV_MONGO_DATABASE)
	if !exists {
		logger.Fatal("mongo database name must be set")
	}

	mongoConn, err := fb.NewMongoDBConn(mongoUri, database)
	if err != nil {
		logger.Fatal("failed establishing connection",
			zap.String("uri", mongoUri),
			zap.Error(err))
	} else {
		logger.Info("connection with mongodb cluster established")
	}

	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, logger)
	directoryServer := dir.NewDirectoryServer(directoryApp, logger, authHeader)

	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	fileApp := file.NewFileApplication(fileRepo, logger)
	fileServer := file.NewFileServer(fileApp, authHeader, logger)

	directoryHandler := dir.NewDirectoryEventHandler(directoryApp, logger)
	if err := fileApp.Subscribe(directoryHandler); err != nil {
		logger.Fatal("failed subscribing directory to files events",
			zap.Error(err))
	}

	grpcSrv := grpc.NewServer()
	proto.RegisterDirectoryServer(grpcSrv, directoryServer)
	proto.RegisterFileServer(grpcSrv, fileServer)

	if err := grpcSrv.Serve(lis); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("no dotenv file has been found",
			zap.Error(err))
	}

	if port, exists := os.LookupEnv(ENV_SERVICE_PORT); exists {
		servicePort = port
	}

	if netw, exists := os.LookupEnv(ENV_SERVICE_NETW); exists {
		serviceNetw = netw
	}

	if header, exists := os.LookupEnv(ENV_AUTH_HEADER); exists {
		authHeader = header
	}

	lis, err := net.Listen(serviceNetw, servicePort)
	if err != nil {
		logger.Panic("failed to listen: %v",
			zap.Error(err))
	}

	logger.Info("server ready to accept connections",
		zap.String("address", servicePort))

	startServer(lis, logger)
}
