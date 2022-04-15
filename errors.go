package filebrowser

import "errors"

var (
	ErrUnknown          = errors.New("E001")
	ErrNotFound         = errors.New("E002")
	ErrNotAvailable     = errors.New("E003")
	ErrUnauthorized     = errors.New("E004")
	ErrInvalidHeader    = errors.New("E007")
	ErrWrongCredentials = errors.New("E008")
	ErrAlreadyExists    = errors.New("E010")
)
