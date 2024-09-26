package impl

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/m-horky/insights-client-next/api/inventory"
	"github.com/m-horky/insights-client-next/internal"
)

// getCurrentInventoryHost tries to fetch a host entry from Inventory.
//
// Error is returned when machine-id file doesn't exist or when Inventory
// returns no host.
func getCurrentInventoryHost() (*inventory.Host, internal.IError) {
	insightsClientID, err := os.ReadFile(internal.MachineIDFilePath)
	if os.IsNotExist(err) {
		slog.Debug("host is not registered, machine-id does not exist")
		return nil, internal.NewError(inventory.ErrNoHost, err, "This host is not registered.")
	}
	if err != nil {
		slog.Debug("host is not registered, machine-id file not readable", slog.String("error", err.Error()))
		return nil, internal.NewError(inventory.ErrNoHost, err, "This host is not registered.")
	}

	return inventory.GetHost(strings.TrimSpace(string(insightsClientID)))
}

func RunCheckIn(input *Input) internal.IError {
	Spinner.Maybe(input, "Fetching host record from Inventory.")
	_, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil && err.Is(inventory.ErrNoHost) {
		return err
	}
	if err != nil {
		return err
	}

	Spinner.Maybe(input, "Updating host record in Inventory.")
	err = inventory.CheckIn()
	Spinner.Stop()
	if err != nil && err.Is(inventory.ErrNoHost) {
		fmt.Println("This host is not registered")
	}
	if err != nil {
		return err
	}

	fmt.Println("Checked in.")
	return nil
}

// RunStatus calls Inventory API.
func RunStatus(input *Input) internal.IError {
	Spinner.Maybe(input, "Fetching host record from Inventory.")
	host, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil && err.Is(inventory.ErrNoHost) {
		fmt.Println("This host is not registered.")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println("This host is registered.")
	fmt.Printf("* Insights Client ID:    %s\n", host.InsightsClientID)
	fmt.Printf("* Insights Inventory ID: %s\n", host.InsightsInventoryID)
	fmt.Printf("* Organization ID:       %s\n", host.OrganizationID)
	return nil
}

// RunSetDisplayName calls Inventory API.
func RunSetDisplayName(input *Input) internal.IError {
	if input.SetDisplayNameArgs.Name == "" {
		return internal.NewError(internal.ErrInput, nil, "Display name cannot be empty.")
	}

	Spinner.Maybe(input, "Fetching host record from Inventory.")
	host, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil {
		return err
	}

	Spinner.Maybe(input, "Updating host record in Inventory.")
	err = inventory.UpdateDisplayName(host.InsightsInventoryID, input.SetDisplayNameArgs.Name)
	Spinner.Stop()
	if err != nil {
		return err
	}
	fmt.Println("Display name was updated.")
	return nil
}

// RunSetAnsibleHostname calls Inventory API.
func RunSetAnsibleHostname(input *Input) internal.IError {
	if input.SetAnsibleHostnameArgs.Name == "" {
		return internal.NewError(internal.ErrInput, nil, "Ansible hostname cannot be empty.")
	}

	Spinner.Maybe(input, "Fetching host record from Inventory.")
	host, err := getCurrentInventoryHost()
	Spinner.Stop()
	if err != nil {
		return err
	}

	Spinner.Maybe(input, "Updating host record in Inventory.")
	err = inventory.UpdateDisplayName(host.InsightsInventoryID, input.SetAnsibleHostnameArgs.Name)
	Spinner.Stop()
	if err != nil {
		return err
	}
	fmt.Println("Ansible hostname was updated.")
	return nil
}
