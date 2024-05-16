package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m-horky/insights-client-next/internal/api"
	"log/slog"
	"net/url"
)

var service = api.NewService("api/inventory/v1")

func GetHost(insightsClientID string) (Host, error) {
	slog.Debug("querying HBI for a host")

	params := url.Values{}
	params.Set("insights_id", insightsClientID)
	response, err := service.MakeRequest("GET", "hosts", params, map[string][]string{}, []byte{})
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return Host{}, err
	}

	var hosts Hosts
	if err = json.Unmarshal(response.Data, &hosts); err != nil {
		slog.Error("could not unmarshal response", slog.String("error", err.Error()))
		return Host{}, err
	}
	if len(hosts.Results) == 0 {
		slog.Debug("HBI returned no hosts", slog.String("response", string(response.Data)))
		return Host{}, errors.New("no hosts found")
	}
	if len(hosts.Results) > 1 {
		slog.Debug("HBI returned more hosts than expected", slog.Int("count", len(hosts.Results)))
	}

	return hosts.Results[0], nil
}
