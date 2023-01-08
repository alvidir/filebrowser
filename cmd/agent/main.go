package main

import (
	"context"
	"os"
	"sync"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
	"github.com/alvidir/filebrowser/cmd"
	dir "github.com/alvidir/filebrowser/directory"
	"github.com/alvidir/filebrowser/event"
	"github.com/alvidir/filebrowser/file"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func handleRabbitMqUserEvents(ctx context.Context, bus *event.RabbitMqEventBus, handler *event.UserEventHandler, logger *zap.Logger) error {
	exchange, exists := os.LookupEnv(cmd.ENV_RABBITMQ_USERS_EXCHANGE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", cmd.ENV_RABBITMQ_USERS_EXCHANGE))

		return fb.ErrUnknown
	}

	queue, exists := os.LookupEnv(cmd.ENV_RABBITMQ_USERS_QUEUE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", cmd.ENV_RABBITMQ_USERS_QUEUE))

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
	exchange, exists := os.LookupEnv(cmd.ENV_RABBITMQ_FILES_EXCHANGE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", cmd.ENV_RABBITMQ_FILES_EXCHANGE))

		return fb.ErrUnknown
	}

	queue, exists := os.LookupEnv(cmd.ENV_RABBITMQ_FILES_QUEUE)
	if !exists {
		logger.Error("must be set",
			zap.String("varname", cmd.ENV_RABBITMQ_FILES_QUEUE))

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

	mongoConn := cmd.GetMongoConnection(logger)
	fileRepo := file.NewMongoFileRepository(mongoConn, logger)
	directoryRepo := dir.NewMongoDirectoryRepository(mongoConn, fileRepo, logger)
	directoryApp := dir.NewDirectoryApplication(directoryRepo, fileRepo, logger)

	privateKey := cmd.GetPrivateKey(logger)
	tokenTTL := cmd.GetTokenTTL(logger)
	tokenIssuer := cmd.GetTokenIssuer(logger)
	certSrv := cert.NewCertificateService(privateKey, tokenIssuer, tokenTTL, logger)
	certRepo := cert.NewMongoCertificateRepository(mongoConn, logger)
	certApp := cert.NewCertificateApplication(certRepo, certSrv, logger)

	conn := cmd.GetAmqpConnection(logger)
	defer conn.Close()

	ch := cmd.GetAmqpChannel(conn, logger)
	defer ch.Close()

	fileApp := file.NewFileApplication(fileRepo, directoryApp, certApp, logger)
	userEventHandler := event.NewUserEventHandler(directoryApp, fileApp, logger)
	fileEventHandler := event.NewFileEventHandler(directoryApp, fileApp, certApp, logger)

	eventIssuer := cmd.GetEventIssuer(logger)
	fileEventHandler.DiscardIssuer(eventIssuer)

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
