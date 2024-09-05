package collectors

import (
	"errors"
)

var (
	ErrNoCollector = errors.New("no such collector")
	ErrCollection  = errors.New("could not collect")
)
