package collectors

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/shlex"
	"github.com/gookit/ini/v2"

	"github.com/m-horky/insights-client-next/internal/constants"
)

type Collector struct {
	Name        string
	Version     string
	Env         []string
	Exec        string
	ExecArgs    []string
	ContentType string
}

// GetCollector looks for a collector defined on a system.
func GetCollector(app string) (*Collector, error) {
	collectors, err := LoadCollectors()
	if err != nil {
		return nil, fmt.Errorf("could not load collectors: %w", err)
	}
	for _, collector := range collectors {
		if collector.Name == app {
			return collector, nil
		}
	}
	return nil, fmt.Errorf("unknown collector: %s", app)
}

func LoadCollectors() ([]*Collector, error) {
	return LoadCollectorsFromDirectory(constants.CollectorsDirectory)
}

func LoadCollectorsFromDirectory(directory string) ([]*Collector, error) {
	var collectors []*Collector
	paths, err := os.ReadDir(directory)
	if err != nil {
		slog.Error("could not detect collectors", slog.String("path", directory), slog.Any("err", err))
		return collectors, fmt.Errorf("could not detect collectors: %w", err)
	}

	for _, file := range paths {
		err := ini.LoadFiles(filepath.Join(directory, file.Name()))
		if err != nil {
			slog.Warn("malformed collector", slog.String("file", file.Name()), slog.Any("err", err))
			continue
		}
		var data ini.Section = ini.Data()["collector"]

		malformed := false
		for _, key := range []string{"name", "version", "env", "exec", "content-type"} {
			if _, ok := data[key]; !ok {
				slog.Error("collector is missing key-value pair", slog.String("file", file.Name()), slog.String("key", key))
				malformed = true
			}
		}

		execArgs, err := shlex.Split(data["exec"])
		if err != nil {
			slog.Error("collector's exec is malformed", slog.Any("error", err))
			malformed = true
		}

		envKeys, err := shlex.Split(data["env"])
		if err != nil {
			slog.Error("collector's environment is malformed", slog.Any("error", err))
			malformed = true
		}

		if malformed {
			slog.Debug("skipping malformed collector", slog.String("file", file.Name()))
			continue
		}

		collectors = append(
			collectors,
			&Collector{
				Name:        data["name"],
				Version:     data["version"],
				Env:         envKeys,
				Exec:        execArgs[0],
				ExecArgs:    execArgs[1:],
				ContentType: data["content-type"],
			},
		)
	}

	return collectors, nil
}
