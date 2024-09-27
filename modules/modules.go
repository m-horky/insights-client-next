package modules

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"
)

type Module struct {
	// Name is human- and machine-readable name.
	Name string
	// Version is the package version.
	Version string
	// Env is a list of environment variables that should always be set.
	Env []string
	// Exec is path to a binary which should be executed, followed by flags that should always be set.
	Exec []string
	// Commands is a list valid (nestable) subcommands.
	Commands [][]string
	// ArchiveCommand is executed to perform a collection. It must be part of Commands.
	ArchiveCommand []string
	// ArchiveContentType is used as HTTP Content-Type for uploaded data archive.
	ArchiveContentType string
}

func GetModules() []*Module {
	return []*Module{
		GetAdvisorModule(),
		GetComplianceModule(),
		GetMalwareModule(),
	}
}

func GetModule(name string) (*Module, IError) {
	for _, module := range GetModules() {
		if module.Name == name {
			return module, nil
		}
	}
	return nil, NewError(ErrNoModule, nil, fmt.Sprintf("Module not found: %s", name))
}

func CommandExists(name []string) bool {
	for _, module := range GetModules() {
		for _, cmd := range module.Commands {
			if reflect.DeepEqual(cmd, name) {
				return true
			}
		}
	}
	return false
}

// CreateArchiveDirectory creates a new directory with semi-random name at `parent`
// with permissions 750.
func CreateArchiveDirectory(parent string) (string, IError) {
	directory := path.Join(parent, fmt.Sprintf("archive-%d", time.Now().Unix()))
	if err := os.Mkdir(directory, 0o750); err != nil {
		return "", NewError(ErrRun, err, "Could not prepare archive directory.")
	}
	return directory, nil
}

// RunCommand executes module command.
//
// The shell command is constructed as `.Exec + command + args`.
func (m *Module) RunCommand(command, args []string) IError {
	var stdout, stderr bytes.Buffer
	argv := m.Exec
	argv = append(argv, command...)
	argv = append(argv, args...)

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = m.Env

	slog.Debug(
		"running module command",
		slog.String("name", m.Name),
		slog.String("version", m.Version),
		slog.String("command", strings.Join(argv, " ")),
		slog.String("environment", strings.Join(m.Env, " ")),
	)

	err := cmd.Run()
	if err != nil {
		slog.Error("module command failed", slog.String("error", err.Error()))
		return NewError(
			ErrRun,
			errors.Join(err, errors.New(stderr.String())),
			"Could not run module command.",
		)
	}

	slog.Debug("module command finished")
	return nil
}

// Collect runs the module's collection command.
//
// `directory` has to exist and has to be writable.
func (m *Module) Collect(directory string, args []string) IError {
	if len(m.ArchiveCommand) == 0 {
		return NewError(ErrRun, nil, "Module does not have collection capabilities.")
	}

	args = append(args, fmt.Sprintf("--archive=%s", directory))
	return m.RunCommand(m.ArchiveCommand, args)
}
