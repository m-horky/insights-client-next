package core

import (
	"fmt"
	"github.com/google/shlex"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Archive struct {
	Path        string
	ContentType string
}

func (a Archive) Delete() error {
	slog.Debug("deleting archive directory", slog.String("archive path", a.Path))
	if err := os.RemoveAll(filepath.Dir(a.Path)); err != nil {
		slog.Warn("could not delete archive directory", slog.Any("error", err))
	}
	return nil
}

// NewEmptyArchive creates a directory into which the archive should be placed.
func NewEmptyArchive() (Archive, error) {
	base := "/var/tmp/"

	id, err := uuid.NewUUID()
	if err != nil {
		slog.Error("could not generate UUID for archive", slog.Any("error", err))
		return Archive{}, err
	}

	path := filepath.Join(base, "insights-client-"+id.String())

	err = os.Mkdir(path, 0o600)
	if err != nil {
		slog.Error("could not create a directory for archive", slog.Any("error", err))
		return Archive{}, err
	}
	return Archive{Path: filepath.Join(path, "archive"), ContentType: ""}, nil
}

// NewArchive asks a collector to create an archive.
func NewArchive(collector Collector) (Archive, error) {
	archive, err := NewEmptyArchive()
	if err != nil {
		return Archive{}, err
	}
	archive.ContentType = collector.ContentType

	envvars := map[string]string{
		"ARCHIVE_PATH": archive.Path,
		"PYTHONPATH":   os.Getenv("PYTHONPATH"),
	}

	slog.Debug(
		"asking for archive",
		slog.String(
			"command",
			fmt.Sprintf(
				"ARCHIVE_PATH=%s PYTHONPATH=%s %s",
				envvars["ARCHIVE_PATH"], envvars["PYTHONPATH"], collector.Exec,
			),
		),
	)
	command, err := shlex.Split(collector.Exec)
	if err != nil {
		slog.Error("could not parse collector command", slog.Any("error", err))
		_ = archive.Delete()
		return Archive{}, err
	}
	cmd := exec.Command(command[0], command[1:]...)
	for k, v := range envvars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("could not capture stdout", slog.Any("error", err))
		_ = archive.Delete()
		return Archive{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		slog.Error("could not capture stderr", slog.Any("error", err))
		_ = archive.Delete()
		return Archive{}, err
	}

	err = cmd.Start()
	if err != nil {
		slog.Error("could not start command", slog.Any("error", err))
		_ = archive.Delete()
		return Archive{}, err
	}
	stdout, err := io.ReadAll(stdoutPipe)
	if err != nil {
		slog.Warn("could not read stdout", slog.Any("error", err))
	}
	stderr, err := io.ReadAll(stderrPipe)
	if err != nil {
		slog.Warn("could not read stderr", slog.Any("error", err))
	}
	if err = cmd.Wait(); err != nil {
		slog.Warn(
			"archive was not created",
			slog.Any("error", err.Error()),
			slog.String("stdout", string(stdout)),
			slog.String("stderr", string(stderr)),
		)
		_ = archive.Delete()
		return Archive{}, err
	}
	slog.Debug(
		"archive created",
		slog.String("stdout", string(stdout)),
		slog.String("stderr", string(stderr)),
	)

	return archive, nil
}
