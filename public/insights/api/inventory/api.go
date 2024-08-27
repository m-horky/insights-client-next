package inventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/m-horky/insights-client-next/public/insights/http"
)

var service http.Service

// Init has to be called to set up the API configuration for the service.
func Init(s *http.Service) {
	service = *s
}

// Exists returns an Inventory ID if there is exactly one host record that matches Insights Client ID.
//
// Error is returned if there is no host, or if there are multiple hosts present (which may happen due
// to host duplication issues Inventory has suffered from).
//
// This endpoint does not yet exist in production at the time of writing this implementation.
func Exists(insightsClientID string) (*HostID, error) {
	slog.Debug("querying HBI for a host")

	params := url.Values{"insights_id": []string{insightsClientID}}

	response, err := service.MakeRequest("GET", "host_exists", params, make(map[string][]string), nil)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not contact HBI: %w", err)
	}

	if response.Code == 404 {
		slog.Debug("host does not exist")
		return nil, errors.New("host does not exist")
	}
	if response.Code == 409 {
		slog.Debug("multiple host records exist")
		return nil, errors.New("multiple host records exist")
	}

	var host HostID
	err = json.Unmarshal(response.Data, &host)
	if err != nil {
		slog.Error("could not unmarshal response", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not unmarshal response: %w", err)
	}

	slog.Debug("HBI host ID obtained", slog.String("id", host.InsightsInventoryID))
	return &host, nil
}

// GetHost returns full host record from Inventory.
//
// Error is returned if there is no host, the first host is returned if there are multiple hosts present.
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
		slog.Debug("HBI returned no hosts")
		return nil, errors.New("no hosts found")
	}
	if len(hosts.Results) > 1 {
		slog.Debug("HBI returned more hosts", slog.Int("count", len(hosts.Results)))
	}

	slog.Debug("HBI host obtained")
	return &hosts.Results[0], nil
}

// DeleteHost deletes the host record from Inventory.
func DeleteHost(insightsInventoryID string) error {
	slog.Debug("deleting HBI host")

	response, err := service.MakeRequest("DELETE", fmt.Sprintf("hosts/%s", insightsInventoryID), url.Values{}, make(map[string][]string), nil)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return fmt.Errorf("could not contact HBI: %w", err)
	}

	if response.Code != 200 {
		slog.Error("could not unregister host", slog.Any("raw response", response.Data))
		return fmt.Errorf("could not unregister host, received %d", response.Code)
	}
	return nil
}

// UpdateDisplayName changes the name of the host displayed in Inventory.
func UpdateDisplayName(insightsInventoryID, displayName string) error {
	slog.Debug("updating HBI host's display name", slog.String("new name", displayName))

	endpoint := fmt.Sprintf("hosts/%s", insightsInventoryID)

	body, err := json.Marshal(map[string]string{"display_name": displayName})
	if err != nil {
		slog.Error("could not encode payload", slog.String("error", err.Error()))
		return fmt.Errorf("could not encode payload: %w", err)
	}

	response, err := service.MakeRequest(
		"PATCH",
		endpoint,
		url.Values{},
		map[string][]string{"Content-Type": {"application/json"}},
		bytes.NewBuffer(body),
	)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return fmt.Errorf("could not contact HBI: %w", err)
	}

	if response.Code != 200 {
		slog.Error("could not update host's display name", slog.Any("raw response", response.Data))
		return fmt.Errorf("could not update host's display name, received %d", response.Code)
	}
	return nil
}
