package inventory

import (
	"errors"
)

var (
	ErrNoHost    = errors.New("host does not exist")
	ErrManyHosts = errors.New("multiple hosts exist")
)
