package file

import "context"

type fileCreatedEvent struct {
	ctx    context.Context
	fileId string
	path   string
}

func newFileCreatedEvent(ctx context.Context, fileId, path string) *fileCreatedEvent {
	return &fileCreatedEvent{
		ctx:    ctx,
		fileId: fileId,
		path:   path,
	}
}

func (event *fileCreatedEvent) Context() context.Context {
	return event.ctx
}

func (event *fileCreatedEvent) FileId() string {
	return event.fileId
}

func (event *fileCreatedEvent) Path() string {
	return event.path
}
