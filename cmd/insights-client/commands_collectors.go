package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"

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
	// When doing local collection, we don't need to check for registration
	if arguments.OutputFile == "" && arguments.OutputDir == "" {
		if _, err := internal.GetCurrentInventoryHost(); err != nil {
			return err
		}
	} else {
		slog.Debug("running local collection, skipping registration check")
	}

	// validate that the --output-dir exists, is readable and empty
	if arguments.OutputDir != "" {
		_, err := os.Stat(arguments.OutputDir)
		if os.IsNotExist(err) {
			return app.NewError(app.ErrInput, err, "The specified output directory does not exist.")
		}
		if err != nil {
			return app.NewError(app.ErrInput, err, "The specified output directory cannot be used.")
		}
		dirContent, err := os.ReadDir(arguments.OutputDir)
		if err != nil {
			return app.NewError(app.ErrInput, err, "The specified output directory cannot be used.")
		}
		if len(dirContent) > 0 {
			return app.NewError(app.ErrInput, err, "The specified output directory cannot be used, because it is not empty.")
		}
	}
	// validate that the --output-file does not exist and its parent does exist
	if arguments.OutputFile != "" {
		_, err := os.Stat(arguments.OutputFile)
		if !os.IsNotExist(err) {
			return app.NewError(app.ErrInput, err, "The specified output file cannot be used.")
		}
		_, err = os.Stat(path.Dir(arguments.OutputFile))
		if os.IsNotExist(err) {
			return app.NewError(app.ErrInput, err, "The specified output file cannot be used, because its parent directory does not exist.")
		}
		if err != nil {
			return app.NewError(app.ErrInput, err, "The specified output file cannot be used.")
		}
	}

	archiveDirectory, archiveContentType, err := collectArchive(arguments)
	if err != nil {
		return err
	}
	if arguments.OutputDir == archiveDirectory {
		fmt.Printf("Archive was kept uncompressed in '%s'. Its content type is '%s'.\n", archiveDirectory, archiveContentType)
		return nil
	}
	defer os.RemoveAll(archiveDirectory)

	archiveFile, err := compressArchive(archiveDirectory, arguments)
	if err != nil {
		return err
	}
	if arguments.OutputFile == archiveFile {
		fmt.Printf("Archive was kept as '%s'. Its content type is '%s'.\n", archiveFile, archiveContentType)
		return nil
	}
	defer os.Remove(archiveFile)

	archive := ingress.Archive{ContentType: archiveContentType, Path: archiveFile}
	if err = uploadArchive(archive, arguments); err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}

func runUploadExistingArchive(arguments *Arguments) app.HumanError {
	if _, err := internal.GetCurrentInventoryHost(); err != nil {
		return err
	}

	_, err := os.Stat(arguments.Payload)
	if os.IsNotExist(err) {
		return app.NewError(app.ErrInput, err, "The specified payload does not exist.")
	}
	if err != nil {
		return app.NewError(app.ErrInput, err, "The specified payload cannot be used.")
	}

	archive := ingress.Archive{ContentType: arguments.ContentType, Path: arguments.Payload}
	if err := uploadArchive(archive, arguments); err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}
