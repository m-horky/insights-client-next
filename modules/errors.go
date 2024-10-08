package modules

import (
	"errors"
	"strings"
)

var (
	ErrNoModule = errors.New("no such module")
	ErrRun      = errors.New("could not run module")
)

type IError interface {
	Error() string
	Is(error) bool
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

// Error returns complex message created from both the internal type and external reason.
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

// Is compares an input error with the internal error type.
func (e *Error) Is(err error) bool {
	return errors.Is(err, e.typ)
}

// Human returns the human-readable version of an error.
func (e *Error) Human() string {
	return "Error: " + e.human
}
