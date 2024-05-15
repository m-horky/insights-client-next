package enums

import (
	"fmt"
	"strings"
)

// Format represents output format
type Format uint8

const (
	Human Format = iota
	JSON
)

// ParseFormat transforms string representation of format into Format.
func ParseFormat(format string) (Format, error) {
	format = strings.ToLower(format)
	if format == "human" {
		return Human, nil
	}
	if format == "json" {
		return JSON, nil
	}
	return Human, fmt.Errorf("unknown format: %s", format)
}
