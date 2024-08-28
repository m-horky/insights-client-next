package http

import (
	"errors"
)

var (
	ErrServiceUnreachable = errors.New("could not contact the service")
	ErrBadResponse        = errors.New("bad response from the service")
	ErrParseError         = errors.New("data could not be parsed")
)
