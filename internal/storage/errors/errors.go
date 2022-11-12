package errors

import "errors"

var (
	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrLoginAlreadyExists     = errors.New("login already exists")
)

var (
	ErrRefreshSessionIsNotExists = errors.New("refresh session is not exists")
)

var (
	ErrOrderWasUploadedByCurrentUser = errors.New("the order was uploaded by current user")
	ErrOrderWasUploadedByAnotherUser = errors.New("the order was uploaded by another user")
	ErrOrderIsNotExists              = errors.New("order is not exist")
)
