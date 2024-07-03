package actions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/m-horky/insights-client-next/internal/core"
)

func RunCollector(collectorName string) error {
	collector, err := core.GetCollector(collectorName)
	if err != nil {
		fmt.Printf("Error: could not load collector: %s\n", err.Error())
		return err
	}

	archive, err := core.NewArchive(collector)
	if err != nil {
		fmt.Printf("Error: could not create archive: %s\n", err.Error())
		return err
	}

	// TODO Upload
	fmt.Printf("%#v\n", archive)
	return nil
}

func ListCollectors() error {
	collectors, err := core.LoadCollectors()
	if err != nil {
		fmt.Printf("Error: could not load collectors: %s\n", err.Error())
		return err
	}

	var collectorNames []string
	for _, collector := range collectors {
		collectorNames = append(collectorNames, collector.Name)
	}
	slices.Sort(collectorNames)

	fmt.Print("Available collectors: ")
	fmt.Print(strings.Join(collectorNames, ", "))
	fmt.Print("\n")
	return nil
}
