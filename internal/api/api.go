package api

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/m-horky/insights-client-next/internal/app"
	"github.com/m-horky/insights-client-next/public/insights/api/ingress"
	"github.com/m-horky/insights-client-next/public/insights/api/inventory"
	"github.com/m-horky/insights-client-next/public/insights/http"
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
	if err != nil {
		slog.Debug("host is not registered, machine-id file missing or not readable")
		return nil, fmt.Errorf("host is not registered: %w", err)
	}

	return inventory.GetHost(strings.TrimSpace(string(insightsClientID)))
}
