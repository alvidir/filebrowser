package main

import (
	"context"
	"crypto/ecdsa"
	"os"
	"sync"
	"time"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/event"
	"github.com/alvidir/filebrowser/file"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	ENV_RABBITMQ_USERS_EXCHANGE = "RABBITMQ_USERS_EXCHANGE"
	ENV_RABBITMQ_USERS_QUEUE    = "RABBITMQ_USERS_QUEUE"
	ENV_RABBITMQ_FILES_EXCHANGE = "RABBITMQ_FILES_EXCHANGE"
	ENV_RABBITMQ_FILES_QUEUE    = "RABBITMQ_FILES_QUEUE"
	ENV_RABBITMQ_DSN            = "RABBITMQ_DSN"
	ENV_MONGO_DSN               = "MONGO_DSN"
	ENV_MONGO_DATABASE          = "MONGO_INITDB_DATABASE"
	ENV_TOKEN_TIMEOUT           = "TOKEN_TIMEOUT"
	ENV_JWT_SECRET              = "JWT_SECRET"
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

func getAmqpConnection(logger *zap.Logger) *amqp.Connection {
	addr, exists := os.LookupEnv(ENV_RABBITMQ_DSN)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_DSN))
	}

	conn, err := amqp.Dial(addr)
	if err != nil {
		logger.Fatal("establishing connection",
			zap.String("addr", addr),
			zap.Error(err))
	}

	return conn
}

func getAmqpChannel(conn *amqp.Connection, logger *zap.Logger) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("openning channel",
			zap.Error(err))
	}

	return ch
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

func handleRabbitMqUserEvents(ctx context.Context, bus *event.RabbitMqEventBus, handler *event.UserEventHandler, logger *zap.Logger) error {
	exchange, exists := os.LookupEnv(ENV_RABBITMQ_USERS_EXCHANGE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_USERS_EXCHANGE))

		return fb.ErrUnknown
	}

	queue, exists := os.LookupEnv(ENV_RABBITMQ_USERS_QUEUE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_USERS_QUEUE))

		return fb.ErrUnknown
	}

	if err := bus.QueueBind(exchange, queue); err != nil {
		logger.Error("binding queue",
			zap.String("queue", queue),
			zap.String("exchange", exchange),
			zap.Error(err))

		return fb.ErrUnknown
	}

	return bus.Consume(ctx, queue, handler.OnEvent)
}

func handleRabbitMqFileEvents(ctx context.Context, bus *event.RabbitMqEventBus, handler *event.FileEventHandler, logger *zap.Logger) error {
	exchange, exists := os.LookupEnv(ENV_RABBITMQ_FILES_EXCHANGE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_FILES_EXCHANGE))

		return fb.ErrUnknown
	}

	queue, exists := os.LookupEnv(ENV_RABBITMQ_FILES_QUEUE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_FILES_QUEUE))

		return fb.ErrUnknown
	}

	if err := bus.QueueBind(exchange, queue); err != nil {
		logger.Error("binding queue",
			zap.String("queue", queue),
			zap.String("exchange", exchange),
			zap.Error(err))

		return fb.ErrUnknown
	}

	return bus.Consume(ctx, queue, handler.OnEvent)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("loading dotenv file",
			zap.Error(err))
	}

	mongoConn := getMongoConnection(logger)
	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)
	userEventHandler := event.NewUserEventHandler(directoryApp, fileApp, logger)

	privateKey := getPrivateKey(logger)
	tokenTTL := getTokenTTL(logger)
	certSrv := cert.NewCertificateService(privateKey, tokenTTL, logger)
	certRepo := cert.NewMongoCertificateRepository(mongoConn, logger)
	certApp := cert.NewCertificateApplication(certRepo, certSrv, logger)
	fileEventHandler := event.NewFileEventHandler(directoryApp, fileApp, certApp, logger)

	conn := getAmqpConnection(logger)
	defer conn.Close()

	ch := getAmqpChannel(conn, logger)
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	bus := event.NewRabbitMqEventBus(ch, logger)

	var wg sync.WaitGroup
	wg.Add(2)

	defer wg.Wait()

	go func() {
		defer wg.Done()

		err := handleRabbitMqUserEvents(ctx, bus, userEventHandler, logger)
		if err != nil {
			cancel()
		}
	}()

	go func() {
		defer wg.Done()

		err := handleRabbitMqFileEvents(ctx, bus, fileEventHandler, logger)
		if err != nil {
			cancel()
		}
	}()
}
