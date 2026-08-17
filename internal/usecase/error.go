package usecase

import "fmt"

type ErrorKind string

const (
	ErrValidation   ErrorKind = "validation"
	ErrUnauthorized ErrorKind = "unauthorized"
	ErrForbidden    ErrorKind = "forbidden"
	ErrNotFound     ErrorKind = "not_found"
	ErrConflict     ErrorKind = "conflict"
	ErrInternal     ErrorKind = "internal"
)

type AppError struct {
	Kind    ErrorKind
	Code    string
	Message string
	Details any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func NewError(kind ErrorKind, code, message string) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message}
}

func WrapError(kind ErrorKind, code, message string, err error) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message, Err: err}
}
