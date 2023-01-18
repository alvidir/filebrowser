package main

import (
	"net"
	"os"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	"github.com/alvidir/filebrowser/cmd"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"github.com/alvidir/filebrowser/proto"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func getNetworkListener(logger *zap.Logger) net.Listener {
	if addr, exists := os.LookupEnv(cmd.ENV_SERVICE_ADDR); exists {
		cmd.ServiceAddr = addr
	}

	if netw, exists := os.LookupEnv(cmd.ENV_SERVICE_NETW); exists {
		cmd.ServiceNetw = netw
	}

	lis, err := net.Listen(cmd.ServiceNetw, cmd.ServiceAddr)
	if err != nil {
		logger.Panic("failed to listen: %v",
			zap.Error(err))
	}

	return lis
}

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
	directoryServer := dir.NewDirectoryServer(directoryApp, logger, cmd.UidHeader)

	privateKey := cmd.GetPrivateKey(logger)
	tokenTTL := cmd.GetTokenTTL(logger)
	tokenIssuer := cmd.GetTokenIssuer(logger)
	certSrv := cert.NewCertificateService(privateKey, tokenIssuer, tokenTTL, logger)
	certRepo := cert.NewMongoCertificateRepository(mongoConn, logger)
	certApp := cert.NewCertificateApplication(certRepo, certSrv, logger)
	certServer := cert.NewCertificateServer(certApp, logger, cmd.UidHeader)

	conn := cmd.GetAmqpConnection(logger)
	defer conn.Close()

	ch := cmd.GetAmqpChannel(conn, logger)
	defer ch.Close()

	eventIssuer := cmd.GetEventIssuer(logger)
	fileExchange := cmd.GetFileExchange(logger)
	bus := fb.RabbitMqEventBus{
		Chann:  ch,
		Logger: logger,
	}

	fileBus := file.NewFileEventBus(bus, fileExchange, eventIssuer)

	fileApp := file.NewFileApplication(fileRepo, directoryApp, certApp, logger)
	fileServer := file.NewFileServer(fileApp, certApp, fileBus, cmd.UidHeader, logger)

	grpcServer := grpc.NewServer()
	proto.RegisterDirectoryServer(grpcServer, directoryServer)
	proto.RegisterFileServer(grpcServer, fileServer)
	proto.RegisterCertificateServer(grpcServer, certServer)
	lis := getNetworkListener(logger)

	logger.Info("server ready to accept connections",
		zap.String("address", cmd.ServiceAddr))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}
