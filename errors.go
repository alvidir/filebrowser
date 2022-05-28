package filebrowser

import "errors"

var (
	ErrUnknown          = errors.New("E001")
	ErrNotFound         = errors.New("E002")
	ErrNotAvailable     = errors.New("E003")
	ErrUnauthorized     = errors.New("E004")
	ErrInvalidFormat    = errors.New("E006")
	ErrInvalidHeader    = errors.New("E007")
	ErrWrongCredentials = errors.New("E008")
	ErrAlreadyExists    = errors.New("E010")

	ErrChannelClosed = errors.New("channel closed")
)
