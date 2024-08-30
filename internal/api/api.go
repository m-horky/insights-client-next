package api

import (
	"log/slog"
	"os"
	"strings"

	"github.com/m-horky/insights-client-next/internal/app"
	"github.com/m-horky/insights-client-next/public/insights/http"
	"github.com/m-horky/insights-client-next/public/insights/services/ingress"
	"github.com/m-horky/insights-client-next/public/insights/services/inventory"
)

// Set up all API services.
func init() {
	config := app.GetConfiguration()

	inventory.Init(&http.Service{
		Protocol: config.APIProtocol,
		Hostname: config.APIHost,
		Port:     config.APIPort,
		Path:     "api/inventory/v1",
	})
	ingress.Init(&http.Service{
		Protocol: config.APIProtocol,
		Hostname: config.APIHost,
		Port:     config.APIPort,
		Path:     "api/ingress/v1",
	})
}

func GetCurrentInventoryHost() (*inventory.Host, error) {
	insightsClientID, err := os.ReadFile("/etc/insights-client/machine-id")
	if os.IsNotExist(err) {
		slog.Debug("host is not registered, machine-id does not exist")
		return nil, inventory.ErrNoHost
	}
	if err != nil {
		slog.Debug("host is not registered, machine-id file not readable", slog.String("error", err.Error()))
		return nil, inventory.ErrNoHost
	}

	return inventory.GetHost(strings.TrimSpace(string(insightsClientID)))
}
