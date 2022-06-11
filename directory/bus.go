package directory

import (
	"context"
	"encoding/json"
	"sync"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	EVENT_CREATED = "CREATED"
	ProfilePath   = "profile"
)

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserEvent struct {
	ID   int32  `json:"id"`
	Kind string `json:"kind"`
	Profile
}

type RabbitMqDirectoryBus struct {
	dirApp  *DirectoryApplication
	fileApp *file.FileApplication
	chann   *amqp.Channel
	logger  *zap.Logger
}

func NewRabbitMqDirectoryBus(dirApp *DirectoryApplication, fileApp *file.FileApplication, chann *amqp.Channel, logger *zap.Logger) *RabbitMqDirectoryBus {
	return &RabbitMqDirectoryBus{
		dirApp:  dirApp,
		fileApp: fileApp,
		chann:   chann,
		logger:  logger,
	}
}

func (bus *RabbitMqDirectoryBus) onEvent(ctx context.Context, delivery *amqp.Delivery) {
	event := new(UserEvent)
	if err := json.Unmarshal(delivery.Body, event); err != nil {
		bus.logger.Error("unmarshaling event body",
			zap.ByteString("event_body", delivery.Body),
			zap.Error(err))

		return
	}

	switch kind := event.Kind; kind {
	case EVENT_CREATED:
		bus.logger.Info("handling event",
			zap.String("kind", kind))

		bus.onUserCreatedEvent(ctx, event)

	default:
		bus.logger.Warn("unhandled event",
			zap.String("kind", kind))
	}
}

func (bus *RabbitMqDirectoryBus) onUserCreatedEvent(ctx context.Context, event *UserEvent) {
	if _, err := bus.dirApp.Create(ctx, event.ID); err != nil {
		return
	}

	data, err := json.Marshal(event.Profile)
	if err != nil {
		bus.logger.Error("marshaling event profile",
			zap.Error(err))

		return
	}

	file, err := bus.fileApp.Create(ctx, event.ID, ProfilePath, nil)
	if err != nil {
		return
	}

	if _, err := bus.fileApp.Write(ctx, event.ID, file.Id(), data, nil); err != nil {
		return
	}

}

func (bus *RabbitMqDirectoryBus) Consume(ctx context.Context, queue string) error {
	events, err := bus.chann.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		bus.logger.Error("registering a consumer",
			zap.Error(err))

		return err
	}

	bus.logger.Info("waiting for events",
		zap.String("queue", queue))

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				bus.logger.Error("channel closed",
					zap.String("queue", queue))

				return fb.ErrChannelClosed
			}

			wg.Add(1)
			go func(ctx context.Context, wg *sync.WaitGroup) {
				defer wg.Done()
				bus.onEvent(ctx, &event)
			}(ctx, &wg)

		case <-ctx.Done():
			bus.logger.Warn("context cancelled",
				zap.Error(ctx.Err()))

			return ctx.Err()
		}
	}

}
