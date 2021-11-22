package app

import "errors"

var (
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrLinkNotFound     = errors.New("link is not located in the service")
	ErrEmptyBodyPostReq = errors.New("body can't be empty")
)
