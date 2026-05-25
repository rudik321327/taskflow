package utils

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrAlreadyExists  = errors.New("resource already exists")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrValidation     = errors.New("validation failed")
	ErrConflict       = errors.New("conflict")
	ErrInvalidCreds   = errors.New("invalid credentials")
)
