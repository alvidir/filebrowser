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
	CREATED_EVENT_KIND = 0
)

type UserEvent struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"emit"`
	Kind  int8   `json:"kind"`
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

func (bus *RabbitMqDirectoryBus) onEvent(ctx context.Context, event *amqp.Delivery) {
	userEvent := new(UserEvent)
	if err := json.Unmarshal(event.Body, userEvent); err != nil {
		bus.logger.Error("parsing event body",
			zap.Error(err))

		return
	}

	switch kind := userEvent.Kind; kind {
	case CREATED_EVENT_KIND:
		bus.logger.Info("handling event",
			zap.Int8("kind", kind))

		bus.onUserCreatedEvent(ctx, userEvent)

	default:
		bus.logger.Warn("unhandled event",
			zap.Int8("kind", kind))
	}
}

func (bus *RabbitMqDirectoryBus) onUserCreatedEvent(ctx context.Context, event *UserEvent) {
	if dir, err := bus.app.Create(ctx, event.ID); err == nil {
		bus.logger.Info("directory created",
			zap.String("id", dir.id),
			zap.Int32("user_id", dir.userId))
	} else {
		bus.logger.Error("creating directory",
			zap.Int32("user_id", event.ID),
			zap.Error(err))
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
