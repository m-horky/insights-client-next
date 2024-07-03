package actions

import (
	"fmt"

	"github.com/m-horky/insights-client-next/internal/constants"
	"github.com/m-horky/insights-client-next/internal/core"
)

func PrintVersion() error {
	fmt.Printf("Insights Client: %s\n", constants.Version)

	collectors, err := core.LoadCollectors()
	if err != nil {
		fmt.Printf("Error: could not load collectors: %s", err.Error())
		return err
	}

	fmt.Print("Collectors:\n")
	for _, collector := range collectors {
		fmt.Printf("  %s: %s\n", collector.Name, collector.Version)
	}
	return nil
}
