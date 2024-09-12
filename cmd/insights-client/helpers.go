package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"gopkg.in/yaml.v3"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/app"
	"github.com/m-horky/insights-client-next/collectors"
	"github.com/m-horky/insights-client-next/internal"
)

var spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)

// isRichOutput detects whether we can pretty-print output
// (animated spinners, ...).
func isRichOutput(arguments *Arguments) bool {
	if arguments.Debug {
		return false
	}
	if arguments.Format != internal.Human {
		return false
	}
	return true
}

// writeTimestampFile saves current timestamp into a file.
func writeTimestampFile(path string) error {
	now := time.Now()
	timestamp := now.Format("2006-01-02T15:04:05.999Z07:00")
	return os.WriteFile(path, []byte(timestamp), 0775)
}

func writeInventoryGroup(group string) app.HumanError {
	rawData, err := os.ReadFile("/etc/insights-client/tags.yaml")
	if err != nil {
		return app.NewError(app.ErrInput, err, "Could not open tags file.")
	}

	data := make(map[string]any)
	if err = yaml.Unmarshal(rawData, &data); err != nil {
		return app.NewError(nil, err, "Could not read tags file.")
	}

	data["group"] = group

	newData, err := yaml.Marshal(data)
	if err != nil {
		return app.NewError(nil, err, "Could not update tags file.")
	}
	if err = os.WriteFile("/etc/insights-client/tags.yaml", newData, 0644); err != nil {
		return app.NewError(nil, err, "Could not update tags file.")
	}
	return nil
}

// collectArchive instructs a collector to gather data into a directory.
//
// It starts a spinner if the terminal output allows it.
//
// Depending on CLI arguments, it writes the data into a given directory (if --output-dir is set)
// or it writes it to a default directory location.
//
// Returns the path to the archive directory, collector's content type, and optional collection error.
func collectArchive(collector *collectors.Collector, arguments *Arguments) (string, string, app.HumanError) {
	if isRichOutput(arguments) {
		spin.Suffix = fmt.Sprintf(" waiting for '%s' to collect its data", collector.Name)
		spin.Start()
		defer spin.Stop()
	}

	collector.ExecArgs = append(collector.ExecArgs, arguments.CollectorOptions...)

	if arguments.OutputDir != "" {
		archiveDirectory, err := collector.CollectToDirectory(arguments.OutputDir)
		return archiveDirectory, collector.ContentType, err
	}
	archiveDirectory, err := collector.Collect()
	return archiveDirectory, collector.ContentType, err
}

// compressArchive calls methods that compress a directory into an archive.
//
// It starts a spinner if the terminal output allows it.
//
// Depending on CLI arguments, it creates an archive at given path (if --output-file is set)
// or it writes it to a default archive location.
//
// Returns the patch to the archive file.
func compressArchive(archiveDirectory string, arguments *Arguments) (string, app.HumanError) {
	if isRichOutput(arguments) {
		spin.Suffix = " compressing archive"
		spin.Start()
		defer spin.Stop()
	}

	if arguments.OutputFile != "" {
		return internal.CompressDirectoryToPath(archiveDirectory, arguments.OutputFile)
	}
	return internal.CompressDirectory(archiveDirectory)
}

// uploadArchive makes a request to Ingress service.
//
// It starts a spinner if the terminal output allows it.
func uploadArchive(archive ingress.Archive, arguments *Arguments) app.HumanError {
	if isRichOutput(arguments) {
		spin.Suffix = " uploading archive"
		spin.Start()
		defer spin.Stop()
	}
	_, err := ingress.UploadArchive(archive)
	return err
}

// unregisterLocally deletes local cache files.
//
// It does not call the Insights API.
func unregisterLocally() app.HumanError {
	dotRegistered := "/etc/insights-client/.registered"
	machineId := "/etc/insights-client/machine-id"
	dotUnregistered := "/etc/insights-client/.unregistered"
	wasRegistered := false

	for _, fileToDelete := range []string{dotRegistered, machineId} {
		if err := os.Remove(fileToDelete); err != nil {
			if !os.IsNotExist(err) {
				slog.Error(fmt.Sprintf("could not remove %s", fileToDelete), slog.String("error", err.Error()))
			}
		} else {
			slog.Debug(fmt.Sprintf("removed %s", fileToDelete))
			wasRegistered = true
		}
	}

	if wasRegistered {
		if err := writeTimestampFile(dotUnregistered); err != nil {
			slog.Error(fmt.Sprintf("could not create %s", dotUnregistered), slog.String("error", err.Error()))
		} else {
			slog.Debug(fmt.Sprintf("created %s", dotUnregistered))
		}
		return nil
	} else {
		return app.NewError(nil, nil, "This host was not registered.")
	}
}
