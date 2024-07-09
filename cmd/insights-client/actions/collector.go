package actions

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/m-horky/insights-client-next/internal/api/ingress"
	"github.com/m-horky/insights-client-next/internal/core"
)

func RunCollector(collectorName string) error {
	collector, err := core.GetCollector(collectorName)
	if err != nil {
		fmt.Printf("Error: could not load collector: %s\n", err.Error())
		return err
	}

	//slog.Info("collecting canonical facts")
	//_, err := system.GetCanonicalFacts()
	//if err != nil {
	//	slog.Error("could not get canonical facts", slog.Any("error", err))
	//	return err
	//}

	slog.Info("collecting archive")
	archive, err := core.NewArchive(*collector)
	if err != nil {
		fmt.Printf("Error: could not create archive: %s\n", err.Error())
		return err
	}

	slog.Info("uploading archive")
	_, err = ingress.UploadArchive(*archive)
	if err != nil {
		slog.Error("could not upload archive", slog.Any("error", err))
	}

	slog.Info("archive uploaded")
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
