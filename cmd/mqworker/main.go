package main

import (
	"context"
	"os"

	fb "github.com/alvidir/filebrowser"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/file"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	ENV_RABBITMQ_BUS   = "RABBITMQ_BUS"
	ENV_RABBITMQ_DSN   = "RABBITMQ_DSN"
	ENV_MONGO_DSN      = "MONGO_DSN"
	ENV_MONGO_DATABASE = "MONGO_INITDB_DATABASE"
	EXCHANGE_TYPE      = "fanout"
	QUEUE_NAME         = "filebrowser"
)

func startWorker(ctx context.Context, conn *amqp.Connection, name string, logger *zap.Logger) error {
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

	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)
	fileApp := file.NewFileApplication(fileRepo, directoryApp, logger)

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("openning channel",
			zap.Error(err))

		return err
	}

	defer ch.Close()

	if err = ch.ExchangeDeclare(
		name,          // name
		EXCHANGE_TYPE, // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	); err != nil {
		logger.Error("declaring exchange",
			zap.String("name", name),
			zap.Error(err))

		return err
	}

	q, err := ch.QueueDeclare(
		QUEUE_NAME, // name
		false,      // durable
		false,      // delete when unused
		true,       // exclusive
		false,      // no-wait
		nil,        // arguments
	)

	if err != nil {
		logger.Error("declaring a queue",
			zap.String("name", QUEUE_NAME),
			zap.Error(err))

		return err
	}

	if err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		name,   // exchange
		false,
		nil,
	); err != nil {
		logger.Error("binding a queue",
			zap.String("exchange", name),
			zap.String("queue", q.Name),
			zap.Error(err))

		return err
	}

	rabbitmqBus := dir.NewRabbitMqDirectoryBus(directoryApp, fileApp, ch, logger)
	return rabbitmqBus.Consume(ctx, q.Name)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("no dotenv file has been found",
			zap.Error(err))
	}

	addr, exists := os.LookupEnv(ENV_RABBITMQ_DSN)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_DSN))

		return
	}

	name, exists := os.LookupEnv(ENV_RABBITMQ_BUS)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", ENV_RABBITMQ_BUS))

		return
	}

	conn, err := amqp.Dial(addr)
	if err != nil {
		logger.Error("establishing connection",
			zap.String("addr", addr),
			zap.Error(err))

		return
	}

	defer conn.Close()
	logger.Info("connection with rabbitmq cluster established")

	ctx := context.Background()
	startWorker(ctx, conn, name, logger)
}
