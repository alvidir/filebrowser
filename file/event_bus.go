package file

import (
	"encoding/json"

	fb "github.com/alvidir/filebrowser"
	"github.com/streadway/amqp"
)

type FileEventPayload struct {
	UserID    int32  `json:"user_id"`
	AppID     string `json:"app_id"`
	FileName  string `json:"file_name"`
	FileID    string `json:"file_id"`
	Reference string `json:"file_reference"`
	Issuer    string `json:"event_issuer"`
	Kind      string `json:"event_kind"`
}

type FileEventBus struct {
	*fb.RabbitMqEventBus
	issuer   string
	exchange string
}

func (bus *FileEventBus) emit(body FileEventPayload) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	err = bus.Chann().Publish(bus.exchange, "", true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(payload),
	})

	if err != nil {
		return err
	}

	return nil
}

func NewFileEventBus(bus *fb.RabbitMqEventBus, exchange string, issuer string) *FileEventBus {
	return &FileEventBus{
		bus,
		issuer,
		exchange,
	}
}

func (bus *FileEventBus) EmitFileCreated(uid int32, f *File) error {
	body := FileEventPayload{
		Issuer:   bus.issuer,
		UserID:   uid,
		AppID:    f.Metadata()[MetadataAppKey],
		FileName: f.Name(),
		FileID:   f.Id(),
		Kind:     fb.EventKindCreated,
	}

	return bus.emit(body)
}

func (bus *FileEventBus) EmitFileDeleted(uid int32, f *File) error {
	body := FileEventPayload{
		Issuer:   bus.issuer,
		UserID:   uid,
		AppID:    f.Metadata()[MetadataAppKey],
		FileName: f.Name(),
		FileID:   f.Id(),
		Kind:     fb.EventKindDeleted,
	}

	return bus.emit(body)
}
