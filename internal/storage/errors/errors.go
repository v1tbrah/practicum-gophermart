package errors

import "errors"

var (
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrLoginAlreadyExists     = errors.New("login already exists")
)

var (
	ErrRefreshSessionIsNotExists = errors.New("refresh session is not exists")
)
