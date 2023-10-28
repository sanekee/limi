package limi

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrHandleExists         = errors.New("handle already exists")
	ErrUnsupportedOperation = errors.New("unsupported operation")

	ErrNotFound         = errors.New("not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
