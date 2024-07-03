package core

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gookit/ini/v2"

	"github.com/m-horky/insights-client-next/internal/constants"
)

type Collector struct {
	Name        string
	Version     string
	Exec        string
	ContentType string
}

// GetCollector looks for a collector defined on a system.
func GetCollector(app string) (Collector, error) {
	collectors, err := LoadCollectors()
	if err != nil {
		return Collector{}, err
	}
	for _, collector := range collectors {
		if collector.Name == app {
			return collector, nil
		}
	}
	return Collector{}, fmt.Errorf("unknown collector: %s", app)
}

func LoadCollectors() ([]Collector, error) {
	return LoadCollectorsFromDirectory(constants.CollectorsDirectory)
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
			slog.Warn("malformed collector", slog.String("file", file.Name()), slog.Any("err", err))
			continue
		}
		var data ini.Section = ini.Data()["collector"]

		missingKeys := false
		for _, key := range []string{"name", "version", "exec", "content-type"} {
			if _, ok := data[key]; !ok {
				slog.Error("collector is missing key-value pair", slog.String("key", key), slog.String("path", file.Name()))
				missingKeys = true
			}
		}

		if missingKeys {
			slog.Debug("skipping malformed collector", slog.String("file", file.Name()))
			continue
		}

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
