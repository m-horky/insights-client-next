package main

import (
	"github.com/m-horky/insights-client-next/internal"
)

// isRichOutput detects whether we can pretty-print output
// (animated spinners, ...).
func isRichOutput(arguments Arguments) bool {
	if arguments.Debug {
		return false
	}
	if arguments.Format != internal.Human {
		return false
	}
	return true
}
