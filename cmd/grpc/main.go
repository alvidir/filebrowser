package main

import (
	"os"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/cmd"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"github.com/alvidir/filebrowser/proto"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	directoryGrpcService := dir.NewDirectoryGrpcServer(directoryApp, logger, cmd.UidHeader)

	conn := cmd.GetAmqpConnection(logger)
	defer conn.Close()

	ch := cmd.GetAmqpChannel(conn, logger)
	defer ch.Close()

	eventIssuer := cmd.GetEventIssuer(logger)
	fileExchange := cmd.GetFileExchange(logger)
	bus := fb.NewRabbitMqEventBus(ch, logger)

	fileBus := file.NewFileEventBus(bus, fileExchange, eventIssuer)

	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)
	fileGrpcService := file.NewFileGrpcServer(fileApp, fileBus, cmd.UidHeader, logger)

	grpcServer := grpc.NewServer()
	proto.RegisterDirectoryServiceServer(grpcServer, directoryGrpcService)
	proto.RegisterFileServiceServer(grpcServer, fileGrpcService)
	lis := cmd.GetNetworkListener(logger)

	logger.Info("server ready to accept connections",
		zap.String("address", cmd.ServiceAddr))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}
