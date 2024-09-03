package app

import (
	"errors"
	"strings"
)

var (
	ErrInput         = errors.New("bad program input")
	ErrPermissions   = errors.New("bad permissions")
	ErrRegistered    = errors.New("host is registered")
	ErrNotRegistered = errors.New("host is not registered")
)

type HumanError interface {
	Error() string
	Human() string
}

// Error is a complex, translatable error object.
type Error struct {
	typ      error
	original error
	human    string
}

// NewError creates a high-level error object.
//
// `typ` is the high-level error defined by the library itself.
//
// `original` is the originally raised error; may be `nil`.
//
// `human` is human-readable, translatable error message displayed to the user.
func NewError(typ, original error, human string) *Error {
	return &Error{typ: typ, original: original, human: human}
}

func (e *Error) Error() string {
	var messages []string
	if e.typ != nil {
		messages = append(messages, e.typ.Error())
	}
	if e.original != nil {
		messages = append(messages, e.original.Error())
	}
	message := strings.Join(messages, ", ")
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, "\n", "; ")
	return message
}

func (e *Error) Human() string {
	return "Error: " + e.human
}
