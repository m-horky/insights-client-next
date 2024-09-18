package modules

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

// Module manages configuration required for data collection.
type Module struct {
	Name        string
	Version     string
	Env         []string
	Exec        string
	ExecArgs    []string
	ContentType string
}

// GetModules gathers all available data collectors.
func GetModules() []*Module {
	return []*Module{
		GetAdvisorModule(),
	}
}

// GetDefaultModule returns the collector that should be run by default.
func GetDefaultModule() *Module {
	return GetAdvisorModule()
}

// GetModule filters available collectors by name.
func GetModule(name string) (*Module, IError) {
	for _, module := range GetModules() {
		if module.Name == name {
			return module, nil
		}
	}
	return nil, NewError(ErrNoModule, nil, fmt.Sprintf("Module not found: %s.", name))
}

func (m *Module) Run(action string) IError {
	var stdout, stderr bytes.Buffer
	args := []string{action}
	args = append(args, m.ExecArgs...)
	cmd := exec.Command(m.Exec, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = m.Env

	slog.Debug(
		"running module",
		slog.String("name", m.Name),
		slog.String("version", m.Version),
		slog.String("exec", fmt.Sprintf("%s %s", m.Exec, strings.Join(m.ExecArgs, " "))),
		slog.String("environment", strings.Join(cmd.Env, " ")),
	)

	err := cmd.Run()
	if err != nil {
		return NewError(
			ErrRun,
			errors.Join(err, errors.New(stderr.String())),
			"Could not run module.",
		)
	}

	return nil
}

// Collect invokes the data collector.
//
// Returns path to a directory with collected data.
func (m *Module) Collect() (string, IError) {
	archiveDirectory := path.Join(ArchiveDirectory, fmt.Sprintf("archive-%d", time.Now().Unix()))
	if err := os.Mkdir(archiveDirectory, 0o750); err != nil {
		return "", NewError(ErrRun, err, "Could not prepare archive directory.")
	}

	return m.CollectToDirectory(archiveDirectory)
}

// CollectToDirectory invokes the module's data collector.
//
// Takes and returns a path to a directory with collected data.
func (m *Module) CollectToDirectory(archiveDirectory string) (string, IError) {
	var stdout, stderr bytes.Buffer
	args := []string{"collect"}
	args = append(args, m.ExecArgs...)
	cmd := exec.Command(m.Exec, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = m.Env
	cmd.Env = append(cmd.Env, fmt.Sprintf("ARCHIVE=%s", archiveDirectory))

	slog.Debug(
		"running module",
		slog.String("name", m.Name),
		slog.String("version", m.Version),
		slog.String("exec", fmt.Sprintf("%s %s", m.Exec, strings.Join(m.ExecArgs, " "))),
		slog.String("environment", strings.Join(cmd.Env, " ")),
	)

	err := cmd.Run()
	if err != nil {
		return "", NewError(
			ErrRun,
			errors.Join(err, errors.New(stderr.String())),
			"Could not run module's collector.",
		)
	}

	return archiveDirectory, nil
}
