package service

import (
	platformservice "perfect-pic-server/internal/common"
)

type AuthErrorCode = platformservice.ErrorCode
type AuthError = platformservice.ServiceError

const (
	AuthErrorValidation   = AuthErrorCode(platformservice.ErrorCodeValidation)
	AuthErrorUnauthorized = AuthErrorCode(platformservice.ErrorCodeUnauthorized)
	AuthErrorForbidden    = AuthErrorCode(platformservice.ErrorCodeForbidden)
	AuthErrorConflict     = AuthErrorCode(platformservice.ErrorCodeConflict)
	AuthErrorNotFound     = AuthErrorCode(platformservice.ErrorCodeNotFound)
	AuthErrorInternal     = AuthErrorCode(platformservice.ErrorCodeInternal)
)

func newAuthError(code AuthErrorCode, message string) error {
	return platformservice.NewServiceError(platformservice.ErrorCode(code), message)
}

func AsAuthError(err error) (*AuthError, bool) {
	serviceErr, ok := platformservice.AsServiceError(err)
	if !ok {
		return nil, false
	}
	return serviceErr, true
}
