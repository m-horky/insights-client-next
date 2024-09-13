package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/internal"
)

func runDisplayName(arguments *Arguments) internal.IError {
	host, err := internal.GetCurrentInventoryHost()
	if err != nil {
		return err
	}

	displayName := arguments.DisplayName
	if arguments.ResetDisplayName {
		hostname, osErr := os.Hostname()
		if osErr != nil {
			slog.Error("could not determine hostname", slog.String("error", osErr.Error()))
			return internal.NewError(nil, osErr, "Could not reset display name.")
		}
		displayName = hostname
	}

	err = inventory.UpdateDisplayName(host.InsightsInventoryID, displayName)
	if err != nil {
		return err
	}
	if arguments.ResetDisplayName {
		fmt.Println("Display name reset.")
	} else {
		fmt.Println("Display name updated.")
	}
	return nil
}

func runAnsibleHostname(arguments *Arguments) internal.IError {
	host, err := internal.GetCurrentInventoryHost()
	if err != nil {
		fmt.Println("Error: Could not get current Inventory host.")
		return err
	}

	ansibleHostname := arguments.AnsibleHost
	if arguments.ResetAnsibleHost {
		hostname, osErr := os.Hostname()
		if osErr != nil {
			slog.Error("could not determine hostname", slog.String("error", osErr.Error()))
			return internal.NewError(nil, osErr, "Could not reset ansible hostname.")
		}
		ansibleHostname = hostname
	}

	err = inventory.UpdateAnsibleHostname(host.InsightsInventoryID, ansibleHostname)
	if err != nil {
		return err
	}
	if arguments.ResetAnsibleHost {
		fmt.Println("Ansible hostname reset.")
	} else {
		fmt.Println("Ansible hostname updated.")
	}
	return nil
}

func runGroup(arguments *Arguments) internal.IError {
	slog.Debug("Setting Inventory group", slog.String("value", arguments.Group))

	if err := writeInventoryGroup(arguments.Group); err != nil {
		return err
	}

	fmt.Println("Tags file updated.")
	return nil

	// TODO Run minimal collection with just the tags data in the archive?
}
