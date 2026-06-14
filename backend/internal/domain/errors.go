package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation error")
	ErrArchivedOnly = errors.New("resource must be archived before deletion")
)

// AppError wraps a sentinel error with a human-readable message and optional field name.
type AppError struct {
	Sentinel error
	Message  string
	Field    string
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Sentinel }

func NotFound(msg string) *AppError {
	return &AppError{Sentinel: ErrNotFound, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Sentinel: ErrUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Sentinel: ErrForbidden, Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{Sentinel: ErrConflict, Message: msg}
}

func ValidationErr(field, msg string) *AppError {
	return &AppError{Sentinel: ErrValidation, Message: msg, Field: field}
}

func ArchivedOnly(msg string) *AppError {
	return &AppError{Sentinel: ErrArchivedOnly, Message: msg}
}
