package filebrowser

import "errors"

var (
	ErrUnknown       = errors.New("E001")
	ErrNotFound      = errors.New("E002")
	ErrAlreadyExists = errors.New("E003")
)
