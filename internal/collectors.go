package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func CompressDirectory(directory string) (string, IError) {
	archive := fmt.Sprintf("%s.tar.xz", directory)

	return CompressDirectoryToPath(directory, archive)
}

func CompressDirectoryToPath(directory, archive string) (string, IError) {
	var stderr bytes.Buffer
	cmd := exec.Command("tar", "--create", "--xz", "--sparse", "--file", archive, directory)
	cmd.Stderr = &stderr

	slog.Debug("compressing archive", slog.String("command", strings.Join(cmd.Args, " ")))

	err := cmd.Run()
	if err != nil {
		return "", NewError(
			nil,
			errors.Join(err, errors.New(stderr.String())),
			"Could not compress archive.",
		)
	}

	stat, err := os.Stat(archive)
	if err != nil {
		return "", NewError(
			nil,
			err,
			"Could not analyze generated archive.",
		)
	}
	slog.Debug("archive created", slog.String("path", archive), slog.Int64("size (kB)", stat.Size()/1000))

	return archive, nil
}
