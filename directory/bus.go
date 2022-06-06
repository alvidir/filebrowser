package directory

import (
	"context"
	"encoding/json"
	"sync"

	fb "github.com/alvidir/filebrowser"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	EVENT_CREATED = "CREATED"
)

type UserEvent struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Kind  string `json:"kind"`
}

type RabbitMqDirectoryBus struct {
	app    *DirectoryApplication
	chann  *amqp.Channel
	logger *zap.Logger
}

func NewRabbitMqDirectoryBus(app *DirectoryApplication, chann *amqp.Channel, logger *zap.Logger) *RabbitMqDirectoryBus {
	return &RabbitMqDirectoryBus{
		app:    app,
		chann:  chann,
		logger: logger,
	}
}

func (bus *RabbitMqDirectoryBus) onEvent(ctx context.Context, delivery *amqp.Delivery) {
	event := new(UserEvent)
	if err := json.Unmarshal(delivery.Body, event); err != nil {
		bus.logger.Error("parsing event body",
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
	_, err := bus.app.Create(ctx, event.ID)
	if err != nil {
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

	for {
		select {
		case event, ok := <-events:
			if !ok {
				bus.logger.Error("channel closed",
					zap.String("queue", queue))

				wg.Wait()
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

			wg.Wait()
			return ctx.Err()
		}
	}

}
