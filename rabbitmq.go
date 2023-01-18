package filebrowser

import (
	"context"
	"sync"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

const (
	EventKindCreated = "created"
	EventKindDeleted = "deleted"
	ExchangeType     = "fanout"
)

type EventHandler func(ctx context.Context, body []byte)

type RabbitMqEventBus struct {
	Chann  *amqp.Channel
	Logger *zap.Logger
}

func (bus *RabbitMqEventBus) QueueBind(exchange, queue string) error {
	if err := bus.Chann.ExchangeDeclare(
		exchange,     // name
		ExchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	); err != nil {
		bus.Logger.Fatal("declaring exchange",
			zap.String("name", exchange),
			zap.Error(err))

		return err
	}

	if _, err := bus.Chann.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		bus.Logger.Fatal("declaring a queue",
			zap.String("name", queue),
			zap.Error(err))

		return err
	}

	if err := bus.Chann.QueueBind(
		queue,    // queue name
		"",       // routing key
		exchange, // exchange name
		false,
		nil,
	); err != nil {
		bus.Logger.Fatal("binding a queue",
			zap.String("exchange", exchange),
			zap.String("queue", queue),
			zap.Error(err))

		return err
	}

	return nil
}

func (bus *RabbitMqEventBus) Consume(ctx context.Context, queue string, handler EventHandler) error {
	events, err := bus.Chann.Consume(
		queue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if err != nil {
		bus.Logger.Error("registering a consumer",
			zap.Error(err))

		return err
	}

	bus.Logger.Info("waiting for events",
		zap.String("queue", queue))

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				bus.Logger.Error("channel closed",
					zap.String("queue", queue))

				return ErrChannelClosed
			}

			wg.Add(1)
			go func(ctx context.Context, wg *sync.WaitGroup) {
				defer wg.Done()
				handler(ctx, event.Body)
			}(ctx, &wg)

		case <-ctx.Done():
			bus.Logger.Warn("context cancelled",
				zap.Error(ctx.Err()))

			return ctx.Err()
		}
	}
}
