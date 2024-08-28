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

func runDisplayName(arguments *Arguments) error {
	host, err := api.GetCurrentInventoryHost()
	if err != nil {
		fmt.Println("Error: could not get current inventory host.")
		return err
	}

	displayName := arguments.DisplayName
	if arguments.ResetDisplayName {
		displayName, err = os.Hostname()
		if err != nil {
			fmt.Printf("Error: Could not reset display name.")
			slog.Error("could not determine hostname", slog.String("error", err.Error()))
			return err
		}
	}

	err = inventory.UpdateDisplayName(host.InsightsInventoryID, displayName)
	if err != nil {
		fmt.Println("Error: Could not update display name.")
		return err
	}
	if arguments.ResetDisplayName {
		fmt.Println("Display name reset.")
	} else {
		fmt.Println("Display name updated.")
	}
	return nil
}

func runAnsibleHostname(arguments *Arguments) error {
	host, err := api.GetCurrentInventoryHost()
	if err != nil {
		fmt.Println("Error: Could not get current Inventory host.")
		return err
	}

	ansibleHostname := arguments.AnsibleHost
	if arguments.ResetAnsibleHost {
		ansibleHostname, err = os.Hostname()
		if err != nil {
			fmt.Printf("Error: Could not reset Ansible hostname.")
			slog.Error("could not determine hostname", slog.String("error", err.Error()))
			return err
		}
	}

	err = inventory.UpdateAnsibleHostname(host.InsightsInventoryID, ansibleHostname)
	if err != nil {
		fmt.Println("Error: Could not update Ansible hostname.")
		return err
	}
	if arguments.ResetAnsibleHost {
		fmt.Println("Ansible hostname reset.")
	} else {
		fmt.Println("Ansible hostname updated.")
	}
	return nil
}

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
	fmt.Printf("* Account:               %s\n", host.Account)
	return nil
}

func runUnregister() error {
	host, err := api.GetCurrentInventoryHost()
	if errors.Is(err, http.ErrServiceUnreachable) {
		fmt.Println("Error: Could not contact Inventory.")
		return err
	}
	if errors.Is(err, inventory.ErrNoHost) {
		fmt.Println("This host is not registered.")
		return nil
	}

	err = inventory.DeleteHost(host.InsightsInventoryID)
	if err != nil {
		fmt.Println("Error: Could not delete host.")
	}

	if err = os.Remove("/etc/insights-client/.registered"); err != nil {
		slog.Error("Could not remove /etc/insights-client/.registered", slog.String("error", err.Error()))
	} else {
		slog.Debug("deleted /etc/insights-client/.registered")
	}
	if err = os.WriteFile("/etc/insights-client/.unregistered", []byte(""), 0644); err != nil {
		slog.Error("Could not create .unregistered file", slog.String("error", err.Error()))
	} else {
		slog.Debug("created etc/insights-client/.unregistered file")
	}

	// TODO Disable systemd service

	fmt.Println("This host was unregistered.")
	return nil
}

func runCollectorList() error {
	fmt.Println("Available collectors:")
	for _, collector := range collectors.GetCollectors() {
		fmt.Printf("* %s %s\n", collector.Name, collector.Version)
	}
	return nil
}
