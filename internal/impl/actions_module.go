package impl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/modules"
)

func RunModule(input *Input) internal.IError {
	args := input.Args.(ARunModuleArgs)

	Spinner.Maybe(input, "Fetching host record from Inventory.")
	_, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil {
		return err
	}

	module, ok := modules.GetModuleByCommand(args.Command)
	if !ok {
		return internal.NewError(internal.ErrInput, nil, fmt.Sprintf("No module implements command '%s'.", strings.Join(args.Command, " ")))
	}

	archiveDirectory, err := modules.CreateArchiveDirectory(args.ArchiveParent)
	if err != nil {
		return err
	}
	if !args.StopAtDir {
		defer os.RemoveAll(archiveDirectory)
	}
	Spinner.Maybe(input, "Collecting host data.")
	err = module.Collect(archiveDirectory, args.Options)
	Spinner.Stop()
	if err != nil {
		return err
	}
	if args.StopAtDir {
		fmt.Printf("Data have been collected to '%s'. Its content type is %s.\n", archiveDirectory, module.ArchiveContentType)
		return nil
	}

	Spinner.Maybe(input, "Compressing host data.")
	archiveFile, err := internal.CompressDirectoryToPath(archiveDirectory, filepath.Join(args.ArchiveParent, args.ArchiveName+".tar.xz"))
	Spinner.Stop()
	if err != nil {
		return err
	}
	if !args.StopAtFile && !args.StopAtCleanup {
		defer os.Remove(archiveFile)
	}
	if args.StopAtFile {
		fmt.Printf("Data have been collected to '%s'. Its content type is '%s+tar.xz'.\n", archiveFile, module.ArchiveContentType)
		return nil
	}

	Spinner.Maybe(input, "Uploading data archive.")
	_, err = ingress.UploadArchive(
		ingress.Archive{Path: archiveFile, ContentType: module.ArchiveContentType + "+tar.xz"},
	)
	Spinner.Stop()
	if err != nil {
		return err
	}

	if args.StopAtCleanup {
		fmt.Printf("Data archive has been uploaded, and has been kept at '%s'. Its content type is '%s+tar.xz'.\n", archiveFile, module.ArchiveContentType)
	} else {
		fmt.Printf("Data archive has been uploaded.\n")
	}
	return nil
}

func RunUploadLocalArchive(input *Input) internal.IError {
	args := input.Args.(AUploadLocalArchiveArgs)

	Spinner.Maybe(input, "Uploading data archive.")
	_, err := ingress.UploadArchive(
		ingress.Archive{Path: args.Path, ContentType: args.ContentType},
	)
	Spinner.Stop()
	if err != nil {
		return err
	}

	fmt.Println("Data archive has been uploaded.")
	return nil
}
