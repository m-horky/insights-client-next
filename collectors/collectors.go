package collectors

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Collector manages configuration required for data collection.
type Collector struct {
	Name        string
	Version     string
	Env         []string
	Exec        string
	ExecArgs    []string
	ContentType string
}

// GetCollectors gathers all available data collectors.
func GetCollectors() []*Collector {
	return []*Collector{
		GetAdvisorCollector(),
	}
}

// GetDefaultCollector returns the collector that should be run by default.
func GetDefaultCollector() *Collector {
	return GetAdvisorCollector()
}

// GetCollector filters available collectors by name.
func GetCollector(name string) (*Collector, IError) {
	for _, collector := range GetCollectors() {
		if collector.Name == name {
			return collector, nil
		}
	}
	return nil, NewError(ErrNoCollector, nil, fmt.Sprintf("Collector not found: %s.", name))
}

// Collect invokes the data collector.
//
// Returns path to a directory with collected data.
func (c *Collector) Collect() (string, IError) {
	archiveDirectory := path.Join(ArchiveDirectory, fmt.Sprintf("archive-%d", time.Now().Unix()))
	if err := os.Mkdir(archiveDirectory, 0o750); err != nil {
		return "", NewError(ErrCollection, err, "Could not prepare archive directory.")
	}

	return c.CollectToDirectory(archiveDirectory)
}

// CollectToDirectory invokes the data collector.
//
// Takes and returns a path to a directory with collected data.
func (c *Collector) CollectToDirectory(archiveDirectory string) (string, IError) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(c.Exec, c.ExecArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = []string{"LC_ALL=C.UTF-8"}
	cmd.Env = append(cmd.Env, c.Env...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("ARCHIVE=%s", archiveDirectory))

	slog.Debug(
		"running collector",
		slog.String("name", c.Name),
		slog.String("version", c.Version),
		slog.String("exec", fmt.Sprintf("%s %s", c.Exec, strings.Join(c.ExecArgs, " "))),
		slog.String("environment", strings.Join(cmd.Env, " ")),
	)

	err := cmd.Run()
	if err != nil {
		return "", NewError(
			ErrCollection,
			errors.Join(err, errors.New(stderr.String())),
			"Could not run collector.",
		)
	}

	return archiveDirectory, nil
}
