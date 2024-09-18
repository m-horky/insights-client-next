package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	_ "github.com/briandowns/spinner"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/modules"
)

func runModuleList() internal.IError {
	fmt.Println("Available modules:")
	for _, module := range modules.GetModules() {
		fmt.Printf("* %s %s\n", module.Name, module.Version)
	}
	return nil
}

func runModule(arguments *Arguments) internal.IError {
	fmt.Println(arguments.ModuleOptions)
	return nil
}

func runModuleCollect(arguments *Arguments) internal.IError {
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
			return internal.NewError(internal.ErrInput, err, "The specified output directory does not exist.")
		}
		if err != nil {
			return internal.NewError(internal.ErrInput, err, "The specified output directory cannot be used.")
		}
		dirContent, err := os.ReadDir(arguments.OutputDir)
		if err != nil {
			return internal.NewError(internal.ErrInput, err, "The specified output directory cannot be used.")
		}
		if len(dirContent) > 0 {
			return internal.NewError(internal.ErrInput, err, "The specified output directory cannot be used, because it is not empty.")
		}
	}
	// validate that the --output-file does not exist and its parent does exist
	if arguments.OutputFile != "" {
		_, err := os.Stat(arguments.OutputFile)
		if !os.IsNotExist(err) {
			return internal.NewError(internal.ErrInput, err, "The specified output file cannot be used.")
		}
		_, err = os.Stat(path.Dir(arguments.OutputFile))
		if os.IsNotExist(err) {
			return internal.NewError(internal.ErrInput, err, "The specified output file cannot be used, because its parent directory does not exist.")
		}
		if err != nil {
			return internal.NewError(internal.ErrInput, err, "The specified output file cannot be used.")
		}
	}

	collector, err := modules.GetModule(arguments.Module)
	if err != nil {
		return err
	}
	archiveDirectory, archiveContentType, err := collectArchive(collector, arguments)
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

func runUploadExistingArchive(arguments *Arguments) internal.IError {
	if _, err := internal.GetCurrentInventoryHost(); err != nil {
		return err
	}

	_, err := os.Stat(arguments.Payload)
	if os.IsNotExist(err) {
		return internal.NewError(internal.ErrInput, err, "The specified payload does not exist.")
	}
	if err != nil {
		return internal.NewError(internal.ErrInput, err, "The specified payload cannot be used.")
	}

	archive := ingress.Archive{ContentType: arguments.ContentType, Path: arguments.Payload}
	if err := uploadArchive(archive, arguments); err != nil {
		return err
	}

	fmt.Println("Archive uploaded.")
	return nil
}
