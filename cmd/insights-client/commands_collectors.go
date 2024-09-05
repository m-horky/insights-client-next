package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	_ "github.com/briandowns/spinner"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/app"
	"github.com/m-horky/insights-client-next/collectors"
	"github.com/m-horky/insights-client-next/internal"
)

func runCollectorList() app.HumanError {
	fmt.Println("Available collectors:")
	for _, collector := range collectors.GetCollectors() {
		fmt.Printf("* %s %s\n", collector.Name, collector.Version)
	}
	return nil
}

func runCollector(arguments Arguments) app.HumanError {
	collector, err := collectors.GetCollector(arguments.Collector)
	if err != nil {
		return err
	}

	//if _, err := internal.GetCurrentInventoryHost(); err != nil {
	//	return err
	//}

	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	if isRichOutput(arguments) {
		spin.Suffix = fmt.Sprintf(" waiting for '%s' to collect its data", collector.Name)
		spin.Start()
	}
	archiveDirectory, err := collector.Collect()
	if isRichOutput(arguments) {
		spin.Stop()
	}
	if err != nil {
		return err
	}
	defer os.RemoveAll(archiveDirectory)

	if isRichOutput(arguments) {
		spin.Suffix = " compressing archive"
		spin.Start()
	}
	archiveFile, err := internal.CompressDirectory(archiveDirectory)
	if isRichOutput(arguments) {
		spin.Stop()
	}
	if err != nil {
		return err
	}
	defer os.Remove(archiveFile)

	archive := ingress.Archive{ContentType: collector.ContentType, Path: archiveFile}

	if isRichOutput(arguments) {
		spin.Suffix = " uploading archive"
		spin.Start()
	}
	_, err = ingress.UploadArchive(archive)
	if isRichOutput(arguments) {
		spin.Stop()
	}
	if err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}
