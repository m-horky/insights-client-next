package actions

import (
	"fmt"

	"github.com/m-horky/insights-client-next/internal/collectors"
	"github.com/m-horky/insights-client-next/internal/constants"
)

func PrintVersion() error {
	fmt.Printf("Insights Client: %s\n", constants.Version)

	collectors, err := collectors.LoadCollectors()
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
