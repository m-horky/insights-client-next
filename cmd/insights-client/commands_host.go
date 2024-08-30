package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/internal/api"
	"github.com/m-horky/insights-client-next/public/collectors"
	"github.com/m-horky/insights-client-next/public/insights/http"
	"github.com/m-horky/insights-client-next/public/insights/services/inventory"
)

func runStatus() error {
	host, err := api.GetCurrentInventoryHost()
	if errors.Is(err, http.ErrServiceUnreachable) {
		fmt.Println("Error: Could not contact Inventory.")
		return err
	}
	if errors.Is(err, inventory.ErrNoHost) {
		fmt.Println("Error: This host is not registered.")
		return nil
	}
	if err != nil {
		fmt.Println("Error: Could not get registration status.")
		return err
	}

	fmt.Println("This host is registered.")
	fmt.Printf("* Insights Client ID:    %s\n", host.InsightsClientID)
	fmt.Printf("* Insights Inventory ID: %s\n", host.InsightsInventoryID)
	fmt.Printf("* Organization ID:       %s\n", host.OrganizationID)
	return nil
}

func runRegister(arguments *Arguments) error {
	host, err := api.GetCurrentInventoryHost()
	if errors.Is(err, http.ErrServiceUnreachable) {
		fmt.Println("Error: Could not contact Inventory.")
		return err
	}
	if host != nil {
		fmt.Println("Error: This host is already registered.")
		return nil
	}

	// TODO Read machine-id from RHSM identity certificate
	if err = os.WriteFile("/etc/insights-client/machine-id", []byte("34d83e98-6818-414d-924d-26d71cbc617a"), 0755); err != nil {
		slog.Error("could not create machine-id file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created /etc/insights-client/machine-id")
	}

	arguments.Collector = collectors.GetDefaultCollector().Name
	// TODO This is awful hack, abstract most of the collection into a separate method :)
	err = runCollector(*arguments)
	if err != nil {
		return err
	}

	// delete .unregistered
	if err = os.Remove("/etc/insights-client/.unregistered"); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove /etc/insights-client/.unregistered", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted /etc/insights-client/.unregistered")
	}

	// write .registered
	if err = writeTimestampFile("/etc/insights-client/.registered"); err != nil {
		slog.Error("could not create .registered file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created /etc/insights-client/.registered file")
	}

	// TODO Write /etc/rhsm/facts/insights.facts

	fmt.Println("This host is now registered. Visit https://console.redhat.com/insights/ to see the Red Hat Insights console.")

	return nil
}

// Unregister the host.
//
// Deletes the host from Inventory and deletes local files.
func runUnregister() error {
	host, err := api.GetCurrentInventoryHost()
	if errors.Is(err, http.ErrServiceUnreachable) {
		fmt.Println("Error: Could not contact Inventory.")
		return err
	}

	wasUnregistered := false
	// delete from Inventory
	if host != nil {
		err = inventory.DeleteHost(host.InsightsInventoryID)
		if err != nil {
			fmt.Println("Error: Could not delete host.")
		}
		wasUnregistered = true
	}

	// delete .registered
	if err = os.Remove("/etc/insights-client/.registered"); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove /etc/insights-client/.registered", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted /etc/insights-client/.registered")
		wasUnregistered = true
	}

	// delete machine-id
	if err = os.Remove("/etc/insights-client/machine-id"); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("could not remove /etc/insights-client/machine-id", slog.String("error", err.Error()))
		}
	} else {
		slog.Debug("deleted /etc/insights-client/machine-id")
		wasUnregistered = true
	}

	// write .unregistered
	if wasUnregistered {
		if err = writeTimestampFile("/etc/insights-client/.unregistered"); err != nil {
			slog.Error("could not create .unregistered file", slog.String("error", err.Error()))
		} else {
			slog.Debug("created etc/insights-client/.unregistered file")
		}
	}

	if wasUnregistered {
		fmt.Println("This host was unregistered.")
	} else {
		fmt.Println("This host is not registered.")
	}

	return nil
}
