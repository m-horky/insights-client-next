package system

import (
	"errors"
	"log/slog"
	"os"

	"github.com/m-horky/insights-client-next/internal/api/inventory"
)

func GetInventoryHost() (*inventory.Host, error) {
	slog.Debug("requesting the host from HBI")

	if _, err := os.Stat("/etc/insights-client/machine-id"); err != nil {
		slog.Debug("machine-id file does not exist, host is definitely not registered")
		return nil, errors.New("system is not registered")
	}

	machineID, err := os.ReadFile("/etc/insights-client/machine-id")
	if err != nil {
		slog.Error("machine-id file is not readable")
		return nil, errors.New("machine-id is not readable")
	}

	return inventory.GetHost(string(machineID))
}
