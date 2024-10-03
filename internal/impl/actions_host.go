package impl

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/internal"
	"github.com/m-horky/insights-client-next/modules"
)

// RunRegister performs the collection and writes special files.
func RunRegister(input *Input) internal.IError {
	args := input.Args.(ARegisterArgs)

	if args.Group != "" {
		// TODO Update group
	}

	Spinner.Maybe(input, "Fetching host record from Inventory.")
	host, err := getCurrentInventoryHost()
	Spinner.Stop()
	if host != nil {
		return internal.NewError(internal.ErrRegistered, nil, "This host is already registered.")
	}

	// TODO Read the location from the rhsm configuration file instead
	rhsm, err := internal.ReadRHSMIdentity("/etc/pki/consumer/cert.pem")
	if err != nil {
		return err
	}

	module, err := modules.GetModule(internal.DefaultModuleName)
	if err != nil {
		return err
	}

	archiveDirectory, err := modules.CreateArchiveDirectory(internal.ArchiveDirectoryParentPath)
	if err != nil {
		return err
	}
	Spinner.Maybe(input, "Collecting host data.")
	var options []string
	if args.DisplayName != "" {
		options = append(options, fmt.Sprintf("--display-name=%s", args.DisplayName))
	}
	if args.AnsibleHostname != "" {
		options = append(options, fmt.Sprintf("--ansible-host=%s", args.AnsibleHostname))
	}
	err = module.Collect(archiveDirectory, options)
	Spinner.Stop()
	if err != nil {
		return err
	}
	defer os.RemoveAll(archiveDirectory)

	Spinner.Maybe(input, "Compressing host data.")
	archiveFile, err := internal.CompressDirectoryToPath(archiveDirectory, archiveDirectory+".tar.xz")
	Spinner.Stop()
	if err != nil {
		return err
	}
	defer os.Remove(archiveFile)

	Spinner.Maybe(input, "Uploading data archive.")
	_, err = ingress.UploadArchive(
		ingress.Archive{Path: archiveFile, ContentType: module.ArchiveContentType + "+tar.xz"},
	)
	Spinner.Stop()
	if err != nil {
		return err
	}

	if err = registerLocally(rhsm); err != nil {
		return err
	}

	fmt.Printf(
		"This host is now registered. Visit %s to see the Red Hat Insights console.\n",
		"https://console.redhat.com/insights/",
	)
	return nil
}

// RunUnregister calls Inventory and writes special files.
func RunUnregister(input *Input) internal.IError {
	Spinner.Maybe(input, "Fetching host record from Inventory.")
	host, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil {
		return err
	}

	wasRegistered := false
	err = inventory.DeleteHost(host.InsightsInventoryID)
	if err != nil && !err.Is(inventory.ErrNoHost) {
		return err
	}
	if err == nil {
		wasRegistered = true
	}

	err = unregisterLocally()
	if err != nil {
		wasRegistered = true
	}

	if wasRegistered {
		fmt.Println("This host was unregistered.")
	} else {
		fmt.Println("This host is not registered.")
	}
	return nil
}

// RunSetGroupLocally updates tags.yaml file. It does not perform data collection.
func RunSetGroupLocally(input *Input) internal.IError {
	// TODO
	return internal.NewError(internal.ErrInput, nil, "Archive upload is not implemented.")
}

func RunTestConnection(input *Input) internal.IError {
	// TODO
	return internal.NewError(internal.ErrInput, nil, "Network testing is not implemented.")
}

func RunSupport(input *Input) internal.IError {
	// TODO
	return internal.NewError(internal.ErrInput, nil, "Customer support is not implemented.")
}

// registerLocally creates, updates and deletes local files.
func registerLocally(rhsm string) internal.IError {
	// write /etc/insights-client/machine-id
	if err := os.WriteFile(internal.MachineIDFilePath, []byte(rhsm), 0755); err != nil {
		slog.Error("could not create machine-id file", slog.String("error", err.Error()))
		return internal.NewError(nil, err, "Could not save UUID file.")
	} else {
		slog.Debug("created machine-id file", slog.String("value", rhsm))
	}

	// delete /etc/insights-client/.unregistered
	if err := os.Remove(internal.DotUnregisteredPath); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove .unregistered file", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted .unregistered file")
	}

	// write /etc/insights-client/.registered
	if err := writeTimestampFile(internal.DotRegisteredPath); err != nil {
		slog.Error("could not create .registered file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created .registered file")
	}
	return nil
}

// unregisterLocally creates, updates and deletes local files.
func unregisterLocally() internal.IError {
	wasRegistered := false

	// delete /etc/insights-client/machine-id
	if err := os.Remove(internal.MachineIDFilePath); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove machine-id file", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted machine-id file")
		wasRegistered = true
	}

	// delete /etc/insights-client/.registered
	if err := os.Remove(internal.DotRegisteredPath); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove .registered file", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted .registered file")
		wasRegistered = true
	}

	// write /etc/insights-client/.unregistered
	if err := writeTimestampFile(internal.DotUnregisteredPath); err != nil {
		slog.Error("could not create .unregistered file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created .unregistered file")
	}

	if wasRegistered {
		return nil
	}
	return internal.NewError(nil, nil, "This host was not registered.")
}

func writeTimestampFile(path string) internal.IError {
	timestamp := time.Now().Format(`2006-01-02T15:04:15.999Z07:00`)
	err := os.WriteFile(path, []byte(timestamp), 0755)
	if err != nil {
		return internal.NewError(nil, err, "Could not write timestamp file.")
	}
	return nil
}
