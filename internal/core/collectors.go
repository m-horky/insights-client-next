package core

import (
	"fmt"
	"strings"
)

// TODO Should this be read dynamically from somewhere?

// Collectors is a list of existing data collectors.
var Collectors = []string{"advisor", "compliance", "malware", "ros"}

// VerifyCollector checks if the collector exists.
func VerifyCollector(app string) error {
	for _, collector := range Collectors {
		if collector == app {
			return nil
		}
	}
	return fmt.Errorf(
		"unknown collector: %s; options are: %s",
		app, strings.Join(Collectors, ", "),
	)
}
