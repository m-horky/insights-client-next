package impl

import (
	"fmt"

	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/modules"
)

func RunListModules(input *Input) internal.IError {
	fmt.Println("Available modules:")
	for _, module := range modules.GetModules() {
		fmt.Printf("* %s %s\n", module.Name, module.Version)
	}
	return nil
}
