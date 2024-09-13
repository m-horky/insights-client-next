package inventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/m-horky/insights-client-next/api"
	"github.com/m-horky/insights-client-next/app"
)

var service api.Service

// Init has to be called to set up the API configuration for the service.
func Init(s *api.Service) {
	service = *s
	service.Path = "api/inventory/v1"
}

// GetHost returns full host record from Inventory.
//
// Error is returned if there is no host, the first host is returned if there are multiple hosts present.
func GetHost(insightsClientID string) (*Host, app.HumanError) {
	slog.Debug("querying HBI for a host")

	params := url.Values{}
	params.Set("insights_id", insightsClientID)

	response, err := service.MakeRequest("GET", "hosts", params, map[string][]string{}, nil)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return nil, api.NewError(
			api.ErrServiceUnreachable,
			err,
			nil,
			"Host inventory could not be contacted.",
		)
	}

	if response.Code != 200 {
		slog.Error("HBI request failed", slog.String("raw response", string(response.Data)))
		return nil, api.NewError(
			api.ErrBadResponse,
			nil,
			response,
			"Host inventory returned bad response.",
		)
	}

	var hosts Hosts
	if err := json.Unmarshal(response.Data, &hosts); err != nil {
		slog.Error("could not unmarshal response", slog.String("error", err.Error()))
		return nil, api.NewError(
			api.ErrUnparseable,
			err,
			response,
			"Host inventory response is malformed.",
		)
	}
	if len(hosts.Results) == 0 {
		slog.Debug("HBI returned no hosts")
		return nil, api.NewError(
			ErrNoHost,
			nil,
			response,
			"Host inventory returned no records.",
		)
	}
	if len(hosts.Results) > 1 {
		slog.Debug("HBI returned more hosts", slog.Int("count", len(hosts.Results)))
	}

	slog.Debug("HBI host obtained", slog.String("inventory uuid", hosts.Results[0].InsightsInventoryID))
	return &hosts.Results[0], nil
}

// DeleteHost deletes the host record from Inventory.
func DeleteHost(insightsInventoryID string) app.HumanError {
	slog.Debug("deleting HBI host")

	response, err := service.MakeRequest("DELETE", fmt.Sprintf("hosts/%s", insightsInventoryID), url.Values{}, make(map[string][]string), nil)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return api.NewError(
			api.ErrServiceUnreachable,
			err,
			nil,
			"Host inventory could not be contacted.",
		)
	}

	if response.Code != 200 {
		slog.Error("could not unregister host", slog.Any("raw response", response.Data))
		return api.NewError(
			api.ErrBadResponse,
			nil,
			response,
			"Host could not be unregistered.",
		)
	}
	return nil
}

// UpdateDisplayName changes the name of the host displayed in Inventory.
func UpdateDisplayName(insightsInventoryID, displayName string) app.HumanError {
	slog.Debug("updating HBI host's display name", slog.String("name", displayName))

	endpoint := fmt.Sprintf("hosts/%s", insightsInventoryID)

	body, err := json.Marshal(map[string]string{"display_name": displayName})
	if err != nil {
		slog.Error("could not encode payload", slog.String("error", err.Error()))
		return api.NewError(
			api.ErrUnparseable,
			err,
			nil,
			"Could not encode payload.",
		)
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
		return api.NewError(
			api.ErrServiceUnreachable,
			err,
			nil,
			"Host inventory could not be contacted.",
		)
	}

	if response.Code != 200 {
		slog.Error("could not update host's display name", slog.Any("raw response", response.Data))
		return api.NewError(
			api.ErrBadResponse,
			nil,
			response,
			"Host inventory returned bad response.",
		)
	}
	return nil
}

// UpdateAnsibleHostname changes the name of the host displayed in Inventory.
func UpdateAnsibleHostname(insightsInventoryID, ansibleHostname string) app.HumanError {
	slog.Debug("updating HBI host's display name", slog.String("name", ansibleHostname))

	endpoint := fmt.Sprintf("hosts/%s", insightsInventoryID)

	body, err := json.Marshal(map[string]string{"ansible_host": ansibleHostname})
	if err != nil {
		slog.Error("could not encode payload", slog.String("error", err.Error()))
		return api.NewError(
			api.ErrUnparseable,
			err,
			nil,
			"Could not encode payload.",
		)
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
		return api.NewError(
			api.ErrServiceUnreachable,
			err,
			nil,
			"Host inventory could not be contacted.",
		)
	}

	if response.Code != 200 {
		slog.Error("could not update host's display name", slog.Any("raw response", response.Data))
		return api.NewError(
			api.ErrBadResponse,
			nil,
			response,
			"Host inventory returned bad response.",
		)
	}
	return nil
}
