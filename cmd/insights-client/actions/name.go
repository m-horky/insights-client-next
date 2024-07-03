package actions

import (
	"fmt"

	"github.com/m-horky/insights-client-next/internal/api/inventory"
	"github.com/m-horky/insights-client-next/internal/system"
)

func SetDisplayName(name string) error {
	host, err := system.GetInventoryHost()
	if err != nil {
		fmt.Print("Error: Host is not registered.\n")
		return err
	}

	err = inventory.UpdateDisplayName(host.InsightsInventoryID, name)
	if err != nil {
		fmt.Printf("Error: could not update display name: %s\n", err.Error())
		return err
	}

	fmt.Print("OK: Display name has been updated.\n")
	return nil
}
