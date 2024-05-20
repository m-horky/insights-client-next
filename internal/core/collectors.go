package core

import (
	"fmt"
	"github.com/gookit/ini/v2"
	"github.com/m-horky/insights-client-next/internal/constants"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Collector struct {
	Name        string
	Version     string
	Exec        string
	ContentType string
}

// Collectors is a list of existing data collectors.
var Collectors []Collector

func init() {
	collectors, err := LoadCollectorsFromDirectory(constants.CollectorsDirectory)
	if err != nil {
		slog.Error("failed to load collectors", slog.Any("error", err))
	}
	Collectors = collectors
}

// VerifyCollector checks if the collector exists.
func VerifyCollector(app string) error {
	for _, collector := range Collectors {
		if collector.Name == app {
			return nil
		}
	}
	var names []string
	for _, collector := range Collectors {
		names = append(names, collector.Name)
	}
	return fmt.Errorf(
		"unknown collector: %s; options are: %s",
		app, strings.Join(names, ", "),
	)
}

func LoadCollectorsFromDirectory(directory string) ([]Collector, error) {
	var collectors []Collector
	paths, err := os.ReadDir(directory)
	if err != nil {
		slog.Error("could not detect collectors", slog.String("path", directory), slog.Any("err", err))
		return collectors, err
	}

	for _, file := range paths {
		err := ini.LoadFiles(filepath.Join(directory, file.Name()))
		if err != nil {
			slog.Warn("malformed collector definition", slog.String("path", file.Name()), slog.Any("err", err))
			continue
		}
		data := ini.Data()["collector"]
		collectors = append(
			collectors,
			Collector{
				Name:        data["name"],
				Version:     data["version"],
				Exec:        data["exec"],
				ContentType: data["content-type"],
			},
		)
	}

	return collectors, nil
}
