package actions

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/m-horky/insights-client-next/internal/api/ingress"
	"github.com/m-horky/insights-client-next/internal/collectors"
)

func RunCollector(collectorName string) error {
	collector, err := collectors.GetCollector(collectorName)
	if err != nil {
		fmt.Printf("Error: could not load collector: %s\n", err.Error())
		return fmt.Errorf("could not load collector: %w", err)
	}

	//slog.Info("collecting canonical facts")
	//_, err := system.GetCanonicalFacts()
	//if err != nil {
	//	slog.Error("could not get canonical facts", slog.Any("error", err))
	//	return err
	//}

	slog.Info("collecting archive")
	archive, err := collectors.NewArchive(*collector)
	if err != nil {
		fmt.Printf("Error: could not create archive: %s\n", err.Error())
		return fmt.Errorf("could not create archive: %w", err)
	}

	slog.Info("uploading archive")
	_, err = ingress.UploadArchive(*archive)
	if err != nil {
		slog.Error("could not upload archive", slog.Any("error", err))
		return fmt.Errorf("could not upload archive: %w", err)
	}

	slog.Info("archive uploaded")
	return nil
}

func ListCollectors() error {
	foundCollectors, err := collectors.LoadCollectors()
	if err != nil {
		fmt.Printf("Error: could not load collectors: %s\n", err.Error())
		return err
	}

	var collectorNames []string
	for _, collector := range foundCollectors {
		collectorNames = append(collectorNames, collector.Name)
	}
	slices.Sort(collectorNames)

	fmt.Print("Available collectors: ")
	fmt.Print(strings.Join(collectorNames, ", "))
	fmt.Print("\n")
	return nil
}
