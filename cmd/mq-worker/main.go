package main

import (
	"context"
	"os"

	fb "github.com/alvidir/filebrowser"
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
	ENV_RABBITMQ_DSN            = "RABBITMQ_DSN"
	ENV_MONGO_DSN               = "MONGO_DSN"
	ENV_MONGO_DATABASE          = "MONGO_INITDB_DATABASE"
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

func newAmqpConnection(logger *zap.Logger) *amqp.Connection {
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

func newAmqpChannel(conn *amqp.Connection, logger *zap.Logger) *amqp.Channel {
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("openning channel",
			zap.Error(err))
	}

	return ch
}

func handleRabbitMqUserEvents(ctx context.Context, bus *event.RabbitMqEventBus, handler *event.UserEventHandler, logger *zap.Logger) {
	exchange, exists := os.LookupEnv(ENV_RABBITMQ_USERS_EXCHANGE)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_USERS_EXCHANGE))
	}

	queue, exists := os.LookupEnv(ENV_RABBITMQ_USERS_QUEUE)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_USERS_QUEUE))
	}

	if err := bus.QueueBind(exchange, queue); err != nil {
		logger.Error("binding queue",
			zap.String("queue", queue),
			zap.String("exchange", exchange),
			zap.Error(err))
	}

	bus.Consume(ctx, queue, handler.OnEvent)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("no dotenv file has been found",
			zap.Error(err))
	}

	mongoConn := newMongoConnection(logger)
	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)
	userEventHandler := event.NewUserEventHandler(directoryApp, fileApp, logger)

	conn := newAmqpConnection(logger)
	defer conn.Close()

	ch := newAmqpChannel(conn, logger)
	defer ch.Close()

	ctx := context.Background()
	bus := event.NewRabbitMqEventBus(ch, logger)
	handleRabbitMqUserEvents(ctx, bus, userEventHandler, logger)
}