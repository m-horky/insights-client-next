package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/internal/api"
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
