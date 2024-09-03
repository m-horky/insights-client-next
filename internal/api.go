package internal

import (
	"log/slog"
	"os"
	"strings"

	"github.com/m-horky/insights-client-next/app"

	"github.com/m-horky/insights-client-next/api/inventory"
)

func GetCurrentInventoryHost() (*inventory.Host, app.HumanError) {
	insightsClientID, err := os.ReadFile("/etc/insights-client/machine-id")
	if os.IsNotExist(err) {
		slog.Debug("host is not registered, machine-id does not exist")
		return nil, app.NewError(inventory.ErrNoHost, err, "This host is not registered.")
	}
	if err != nil {
		slog.Debug("host is not registered, machine-id file not readable", slog.String("error", err.Error()))
		return nil, app.NewError(inventory.ErrNoHost, err, "This host is not registered.")
	}

	return inventory.GetHost(strings.TrimSpace(string(insightsClientID)))
}
