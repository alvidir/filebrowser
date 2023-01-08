package event

import (
	"encoding/json"

	"github.com/alvidir/filebrowser/file"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type FileEventBus struct {
	RabbitMqEventBus
	issuer   string
	exchange string
}

func (bus *FileEventBus) emit(body FileEventPayload) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = bus.chann.Publish(bus.exchange, "", true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(payload),
	})

	if err != nil {
		return err
	}

	return nil
}

func NewFileEventBus(issuer string, exchange string, chann *amqp.Channel, logger *zap.Logger) *FileEventBus {
	return &FileEventBus{
		RabbitMqEventBus{
			chann:  chann,
			logger: logger,
		},
		issuer,
		exchange,
	}
}

func (bus *FileEventBus) EmitFileCreated(uid int32, f *file.File) error {
	body := FileEventPayload{
		Issuer:   bus.issuer,
		UserID:   uid,
		AppID:    f.Metadata()[file.MetadataAppKey],
		FileName: f.Name(),
		FileID:   f.Id(),
		Kind:     EventKindCreated,
	}

	return bus.emit(body)
}

func (bus *FileEventBus) EmitFileDeleted(uid int32, f *file.File) error {
	body := FileEventPayload{
		Issuer:   bus.issuer,
		UserID:   uid,
		AppID:    f.Metadata()[file.MetadataAppKey],
		FileName: f.Name(),
		FileID:   f.Id(),
		Kind:     EventKindDeleted,
	}

	return bus.emit(body)
}
