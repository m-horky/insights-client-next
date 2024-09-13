package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/api/ingress"
	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/app"
	"github.com/m-horky/insights-client-next/collectors"
	"github.com/m-horky/insights-client-next/internal"
)

func runStatus() app.HumanError {
	host, err := internal.GetCurrentInventoryHost()
	if err != nil {
		return err
	}

	fmt.Println("This host is registered.")
	fmt.Printf("* Insights Client ID:    %s\n", host.InsightsClientID)
	fmt.Printf("* Insights Inventory ID: %s\n", host.InsightsInventoryID)
	fmt.Printf("* Organization ID:       %s\n", host.OrganizationID)
	return nil
}

func runRegister(arguments *Arguments) app.HumanError {
	host, _ := internal.GetCurrentInventoryHost()
	if host != nil {
		return app.NewError(app.ErrRegistered, nil, "This host is already registered.")
	}

	rhsm, err := internal.ReadRHSMIdentity("/etc/pki/consumer/cert.pem")
	if err != nil {
		return err
	}

	if err := os.WriteFile("/etc/insights-client/machine-id", []byte(rhsm), 0755); err != nil {
		slog.Error("could not create machine-id file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created /etc/insights-client/machine-id", slog.String("value", rhsm))
	}

	if len(arguments.Group) > 0 {
		if err = writeInventoryGroup(arguments.Group); err != nil {
			return err
		}
		fmt.Println("Tags file updated.")
	}

	collector, err := collectors.GetCollector(arguments.Collector)
	if err != nil {
		return err
	}

	// during registration, the Advisor collection reads the values from config
	// and saves them as archive metadata as `/display_name` and `/ansible_host`
	// (see `insights/client/data_collector.py:DataCollector`).
	if arguments.DisplayName != "" {
		collector.ExecArgs = append(collector.ExecArgs, "--display-name="+arguments.DisplayName)
	}
	if arguments.AnsibleHost != "" {
		collector.ExecArgs = append(collector.ExecArgs, "--ansible-host="+arguments.AnsibleHost)
	}

	// run the collection
	arguments.Collector = collectors.GetDefaultCollector().Name
	archiveDirectory, archiveContentType, err := collectArchive(collector, arguments)
	if err != nil {
		return err
	}
	defer os.RemoveAll(archiveDirectory)
	archiveFile, err := compressArchive(archiveDirectory, arguments)
	if err != nil {
		return err
	}
	defer os.Remove(archiveFile)
	if err = uploadArchive(ingress.Archive{ContentType: archiveContentType, Path: archiveFile}, arguments); err != nil {
		return err
	}

	// delete .unregistered
	if err := os.Remove("/etc/insights-client/.unregistered"); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove /etc/insights-client/.unregistered", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted /etc/insights-client/.unregistered")
	}

	// write .registered
	if err := writeTimestampFile("/etc/insights-client/.registered"); err != nil {
		slog.Error("could not create .registered file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created /etc/insights-client/.registered file")
	}

	fmt.Println("This host is now registered. Visit https://console.redhat.com/insights/ to see the Red Hat Insights console.")

	return nil
}

// Unregister the host.
//
// Deletes the host from Inventory and deletes local files.
func runUnregister() app.HumanError {
	host, err := internal.GetCurrentInventoryHost()
	if err != nil {
		return err
	}

	// delete from Inventory
	if host != nil {
		err = inventory.DeleteHost(host.InsightsInventoryID)
		if err != nil && !err.Is(inventory.ErrNoHost) {
			return err
		}
	}

	// delete and update local files
	err = unregisterLocally()
	if err != nil {
		return err
	}
	fmt.Println("This host was unregistered.")
	return nil
}
