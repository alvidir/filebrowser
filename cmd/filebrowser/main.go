package main

import (
	"net"
	"os"
	"strconv"
	"time"

	fb "github.com/alvidir/filebrowser"
	dir "github.com/alvidir/filebrowser/directory"
	file "github.com/alvidir/filebrowser/file"
	perm "github.com/alvidir/filebrowser/permissions"
	proto "github.com/alvidir/filebrowser/proto"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	ENV_SERVICE_ADDR   = "SERVICE_ADDR"
	ENV_SERVICE_NETW   = "SERVICE_NETW"
	ENV_AUTH_HEADER    = "AUTH_HEADER"
	ENV_MONGO_DSN      = "MONGO_DSN"
	ENV_MONGO_DATABASE = "MONGO_DATABASE"
	ENV_REDIS_DSN      = "REDIS_DSN"
	ENV_CACHE_TTL      = "CACHE_TTL"
	ENV_CACHE_SIZE     = "CACHE_SIZE"
)

var (
	serviceAddr = "0.0.0.0:8000"
	serviceNetw = "tcp"
	authHeader  = "X-Auth"
	cacheTTL    = 10 * time.Minute
	cacheSize   = 1024
)

func newMongoConnection(logger *zap.Logger) *mongo.Database {
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

func newRedisCache(logger *zap.Logger) *cache.Cache {
	addr, exists := os.LookupEnv(ENV_REDIS_DSN)
	if !exists {
		logger.Fatal("redis dsn must be set")
	}

	if value, exists := os.LookupEnv(ENV_CACHE_TTL); exists {
		if ttl, err := time.ParseDuration(value); err != nil {
			logger.Fatal("invalid cache ttl",
				zap.String("value", value),
				zap.Error(err))
		} else {
			cacheTTL = ttl
		}
	}

	if value, exists := os.LookupEnv(ENV_CACHE_SIZE); exists {
		if size, err := strconv.Atoi(value); err != nil {
			logger.Fatal("invalid cache size",
				zap.String("value", value),
				zap.Error(err))
		} else {
			cacheSize = size
		}
	}

	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server": addr,
		},
	})

	return cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(cacheSize, cacheTTL),
	})
}

func newFilebrowserGrpcServer(mongoConn *mongo.Database, cache *cache.Cache, logger *zap.Logger) *grpc.Server {
	if header, exists := os.LookupEnv(ENV_AUTH_HEADER); exists {
		authHeader = header
	}

	fileRepo := file.NewMongoFileRepository(mongoConn, logger)

	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	directoryServer := dir.NewDirectoryServer(directoryApp, logger, authHeader)

	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)
	fileServer := file.NewFileServer(fileApp, authHeader, logger)

	permissionsRepo := perm.NewRedisPermissionsRepository(cache, fileRepo, logger)
	permissionsApp := perm.NewPermissionsApplication(permissionsRepo, fileRepo, logger)
	permissionsServer := perm.NewPermissionsServer(permissionsApp, logger, authHeader)

	grpcSrv := grpc.NewServer()
	proto.RegisterDirectoryServer(grpcSrv, directoryServer)
	proto.RegisterFileServer(grpcSrv, fileServer)
	proto.RegisterPermissionsServer(grpcSrv, permissionsServer)

	return grpcSrv
}

func newNetworkListener(logger *zap.Logger) net.Listener {
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
		logger.Warn("no dotenv file has been found",
			zap.Error(err))
	}

	mongoConn := newMongoConnection(logger)
	redisCache := newRedisCache(logger)
	grpcServer := newFilebrowserGrpcServer(mongoConn, redisCache, logger)
	lis := newNetworkListener(logger)

	logger.Info("server ready to accept connections",
		zap.String("address", serviceAddr))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("server terminated with errors",
			zap.Error(err))
	}
}
