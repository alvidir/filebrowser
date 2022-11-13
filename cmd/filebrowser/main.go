package main

import (
	"crypto/ecdsa"
	"net"
	"os"
	"time"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	dir "github.com/alvidir/filebrowser/directory"
	file "github.com/alvidir/filebrowser/file"
	proto "github.com/alvidir/filebrowser/proto"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	ENV_SERVICE_ADDR   = "SERVICE_ADDR"
	ENV_SERVICE_NETW   = "SERVICE_NETW"
	ENV_UID_HEADER     = "UID_HEADER"
	ENV_MONGO_DSN      = "MONGO_DSN"
	ENV_MONGO_DATABASE = "MONGO_DATABASE"
	ENV_REDIS_DSN      = "REDIS_DSN"
	ENV_TOKEN_TIMEOUT  = "TOKEN_TIMEOUT"
	ENV_JWT_SECRET     = "JWT_SECRET"
)

var (
	serviceAddr = "0.0.0.0:8000"
	serviceNetw = "tcp"
	uidHeader   = "X-Uid"
)

func getMongoConnection(logger *zap.Logger) *mongo.Database {
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

	return mongoConn
}

func getPrivateKey(logger *zap.Logger) *ecdsa.PrivateKey {
	secret, exists := os.LookupEnv(ENV_JWT_SECRET)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_JWT_SECRET))
	}

	privateKey, err := cert.ParsePKCS8PrivateKey(secret)
	if err != nil {
		logger.Fatal("parsing PKCS8 private key",
			zap.Error(err))
	}

	return privateKey
}

func getTokenTTL(logger *zap.Logger) *time.Duration {
	value, exists := os.LookupEnv(ENV_TOKEN_TIMEOUT)
	if !exists {
		return nil
	}

	ttl, err := time.ParseDuration(value)
	if err != nil {
		logger.Fatal("invalid token ttl",
			zap.String("value", value),
			zap.Error(err))
	}

	return &ttl
}

func getFilebrowserGrpcServer(mongoConn *mongo.Database, logger *zap.Logger) *grpc.Server {
	if header, exists := os.LookupEnv(ENV_UID_HEADER); exists {
		uidHeader = header
	}

	fileRepo := file.NewMongoFileRepository(mongoConn, logger)

	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	directoryServer := dir.NewDirectoryServer(directoryApp, logger, uidHeader)

	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)
	fileServer := file.NewFileServer(fileApp, uidHeader, logger)

	privateKey := getPrivateKey(logger)
	tokenTTL := getTokenTTL(logger)
	certSrv := cert.NewCertificateService(privateKey, tokenTTL, logger)
	certRepo := cert.NewMongoCertificateRepository(mongoConn, logger)
	certApp := cert.NewCertificateApplication(certRepo, certSrv, logger)
	certServer := cert.NewCertificateServer(certApp, logger, uidHeader)

	grpcSrv := grpc.NewServer()
	proto.RegisterDirectoryServer(grpcSrv, directoryServer)
	proto.RegisterFileServer(grpcSrv, fileServer)
	proto.RegisterCertificateServer(grpcSrv, certServer)

	return grpcSrv
}

func getNetworkListener(logger *zap.Logger) net.Listener {
	if addr, exists := os.LookupEnv(ENV_SERVICE_ADDR); exists {
		serviceAddr = addr
	}

	if netw, exists := os.LookupEnv(ENV_SERVICE_NETW); exists {
		serviceNetw = netw
	}

	lis, err := net.Listen(serviceNetw, serviceAddr)
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

	mongoConn := getMongoConnection(logger)
	grpcServer := getFilebrowserGrpcServer(mongoConn, logger)
	lis := getNetworkListener(logger)

	logger.Info("server ready to accept connections",
		zap.String("address", serviceAddr))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}
