package main

import (
	"fmt"
	"os"

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

func runCollector(arguments *Arguments) app.HumanError {
	if _, err := internal.GetCurrentInventoryHost(); err != nil {
		return err
	}

	archiveDirectory, archiveContentType, err := collectArchive(arguments)
	if err != nil {
		return err
	}
	defer os.RemoveAll(archiveDirectory)

	archiveFile, err := compressArchive(archiveDirectory, arguments)
	if err != nil {
		return err
	}
	defer os.Remove(archiveFile)

	archive := ingress.Archive{ContentType: archiveContentType, Path: archiveFile}
	if err = uploadArchive(archive, arguments); err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}
