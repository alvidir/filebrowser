package main

import (
	"context"
	"os"

	fb "github.com/alvidir/filebrowser"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	ENV_RABBITMQ_USERS_BUS = "RABBITMQ_USERS_BUS"
	ENV_RABBITMQ_DSN       = "RABBITMQ_DSN"
	ENV_MONGO_DSN          = "MONGO_DSN"
	ENV_MONGO_DATABASE     = "MONGO_INITDB_DATABASE"
	EXCHANGE_TYPE          = "fanout"
	QUEUE_NAME             = "filebrowser-users"
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

func initRabbitMQUsersBus(ch *amqp.Channel, logger *zap.Logger) {
	bus, exists := os.LookupEnv(ENV_RABBITMQ_USERS_BUS)
	if !exists {
		logger.Fatal("must be set",
			zap.String("varname", ENV_RABBITMQ_USERS_BUS))
	}

	if err := ch.ExchangeDeclare(
		bus,           // name
		EXCHANGE_TYPE, // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	); err != nil {
		logger.Fatal("declaring exchange",
			zap.String("name", bus),
			zap.Error(err))
	}

	if _, err := ch.QueueDeclare(
		QUEUE_NAME, // name
		false,      // durable
		false,      // delete when unused
		true,       // exclusive
		false,      // no-wait
		nil,        // arguments
	); err != nil {
		logger.Fatal("declaring a queue",
			zap.String("name", QUEUE_NAME),
			zap.Error(err))
	}

	if err := ch.QueueBind(
		QUEUE_NAME, // queue name
		"",         // routing key
		bus,        // exchange name
		false,
		nil,
	); err != nil {
		logger.Fatal("binding a queue",
			zap.String("exchange", bus),
			zap.String("queue", QUEUE_NAME),
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

	mongoConn := newMongoConnection(logger)
	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)

	conn := newAmqpConnection(logger)
	defer conn.Close()

	ch := newAmqpChannel(conn, logger)
	defer ch.Close()

	initRabbitMQUsersBus(ch, logger)

	ctx := context.Background()
	rabbitmqBus := dir.NewRabbitMqDirectoryBus(directoryApp, fileApp, ch, logger)
	rabbitmqBus.Consume(ctx, QUEUE_NAME)
}
