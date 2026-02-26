package service

import "errors"

type ErrorCode string

const (
	ErrorCodeValidation   ErrorCode = "validation"
	ErrorCodeUnauthorized ErrorCode = "unauthorized"
	ErrorCodeForbidden    ErrorCode = "forbidden"
	ErrorCodeConflict     ErrorCode = "conflict"
	ErrorCodeNotFound     ErrorCode = "not_found"
	ErrorCodeInternal     ErrorCode = "internal"
)

type ServiceError struct {
	Code    ErrorCode
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(code ErrorCode, message string) error {
	return &ServiceError{Code: code, Message: message}
}

func NewValidationError(message string) error {
	return NewServiceError(ErrorCodeValidation, message)
}

func NewUnauthorizedError(message string) error {
	return NewServiceError(ErrorCodeUnauthorized, message)
}

func NewForbiddenError(message string) error {
	return NewServiceError(ErrorCodeForbidden, message)
}

func NewConflictError(message string) error {
	return NewServiceError(ErrorCodeConflict, message)
}

func NewNotFoundError(message string) error {
	return NewServiceError(ErrorCodeNotFound, message)
}

func NewInternalError(message string) error {
	return NewServiceError(ErrorCodeInternal, message)
}

func AsServiceError(err error) (*ServiceError, bool) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr, true
	}
	return nil, false
}
