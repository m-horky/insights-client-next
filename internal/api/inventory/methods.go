package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/m-horky/insights-client-next/internal/api"
)

var service = api.NewService("api/inventory/v1")

func GetHost(insightsClientID string) (*Host, error) {
	slog.Debug("querying HBI for a host")

	params := url.Values{}
	params.Set("insights_id", insightsClientID)
	response, err := service.MakeRequest("GET", "hosts", params, map[string][]string{}, nil)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not contact HBI: %w", err)
	}

	var hosts Hosts
	if err = json.Unmarshal(response.Data, &hosts); err != nil {
		slog.Error("could not unmarshal response", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not unmarshal response: %w", err)
	}
	if len(hosts.Results) == 0 {
		slog.Debug("HBI returned no hosts", slog.String("response", string(response.Data)))
		return nil, errors.New("no hosts found")
	}
	if len(hosts.Results) > 1 {
		slog.Debug("HBI returned more hosts than expected", slog.Int("count", len(hosts.Results)))
	}

	slog.Debug("HBI host obtained")
	return &hosts.Results[0], nil
}

func UpdateDisplayName(insightsInventoryID, displayName string) error {
	slog.Debug("updating HBI host's display name", slog.String("new name", displayName))

	params := url.Values{}
	params.Set("display_name", displayName)

	body, err := json.Marshal(map[string]string{"display_name": displayName})
	if err != nil {
		slog.Error("could not encode payload", slog.Any("error", err))
		return fmt.Errorf("could not encode payload: %w", err)
	}

	response, err := service.MakeRequest(
		"PATCH",
		fmt.Sprintf("hosts/%s", insightsInventoryID),
		url.Values{},
		map[string][]string{"Content-Type": {"application/json"}},
		bytes.NewBuffer(body),
	)
	if err != nil {
		slog.Error("could not contact HBI", slog.Any("error", err))
		return fmt.Errorf("could not contact HBI: %w", err)
	}

	if response.Code != 200 {
		slog.Warn("HBI responded with error", slog.String("response", string(response.Data)))
		return fmt.Errorf("could not update host's display name, received status code %d", response.Code)
	}
	return nil
}
