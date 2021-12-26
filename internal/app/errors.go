package app

import "errors"

var (
	ErrMethodNotAllowed  = errors.New("method not allowed")
	ErrLinkNotFound      = errors.New("link is not located in the service")
	ErrLinkAlreadyExists = errors.New("link already exists in the service")
	ErrEmptyBodyPostReq  = errors.New("body can't be empty")
	ErrUserHasNoRecords  = errors.New("user has no records")
	ErrInvalidSignature  = errors.New("invalid signature")
)
