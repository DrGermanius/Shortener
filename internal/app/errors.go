package app

import "errors"

var (
	ErrMethodNotAllowed error
	ErrLinkNotFound     error
)

func init() {
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrLinkNotFound = errors.New("link doesn't locate at the service")
}
