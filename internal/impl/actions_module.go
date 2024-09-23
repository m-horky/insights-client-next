package impl

import (
	"fmt"
	"strings"

	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/modules"
)

func RunListModules(input *Input) internal.IError {
	fmt.Println("Available modules:")

	maxLength := 0
	for _, module := range modules.GetModules() {
		if len(module.Name) > maxLength {
			maxLength = len(module.Name)
		}
	}

	for _, module := range modules.GetModules() {
		fmt.Printf("* %s  %s\n", module.Name+strings.Repeat(" ", maxLength-len(module.Name)), module.Version)
	}
	return nil
}