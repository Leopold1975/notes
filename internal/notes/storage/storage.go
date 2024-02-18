package storage

import (
	"errors"
)

var (
	ErrContextCancelled   = errors.New("context cancelled connection")
	ErrFieldUnspecified   = errors.New("required fields are unspecified")
	ErrDatabaseNotExists  = errors.New("database don't exist")
	ErrNotFound           = errors.New("entity not found")
	ErrNotEnoughArguments = errors.New("not enough arguments in call")
)
