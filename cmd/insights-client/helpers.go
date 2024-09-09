package main

import (
	"fmt"
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
// Returns the path to the archive directory, collector's content type, and optional collection error.
func collectArchive(arguments *Arguments) (string, string, app.HumanError) {
	collector, err := collectors.GetCollector(arguments.Collector)
	if err != nil {
		return "", "", err
	}

	if isRichOutput(arguments) {
		spin.Suffix = fmt.Sprintf(" waiting for '%s' to collect its data", collector.Name)
		spin.Start()
		defer spin.Stop()
	}
	archiveDirectory, err := collector.Collect()
	return archiveDirectory, collector.ContentType, err
}

func compressArchive(archiveDirectory string, arguments *Arguments) (string, app.HumanError) {
	if isRichOutput(arguments) {
		spin.Suffix = " compressing archive"
		spin.Start()
		defer spin.Stop()
	}
	return internal.CompressDirectory(archiveDirectory)
}

func uploadArchive(archive ingress.Archive, arguments *Arguments) app.HumanError {
	if isRichOutput(arguments) {
		spin.Suffix = " uploading archive"
		spin.Start()
		defer spin.Stop()
	}
	_, err := ingress.UploadArchive(archive)
	return err
}
