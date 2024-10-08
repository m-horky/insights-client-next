package internal

import (
	"strings"
)

type Format uint8

const (
	Human Format = iota
	JSON
)

// ParseFormat converts a string into a Format.
//
// If parsing fails, returns `Human` and an error.
func ParseFormat(format string) (Format, IError) {
	switch strings.ToLower(format) {
	case "human":
		return Human, nil
	case "json":
		return JSON, nil
	default:
		return Human, NewError(nil, nil, "Unknown format.")
	}
}

// MustParseFormat converts a string into a Format.
//
// If parsing fails, returns `Human`.
func MustParseFormat(format string) Format {
	switch strings.ToLower(format) {
	case "json":
		return JSON
	default:
		return Human
	}
}
