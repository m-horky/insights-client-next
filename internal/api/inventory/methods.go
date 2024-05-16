package inventory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m-horky/insights-client-next/internal/api"
	"github.com/m-horky/insights-client-next/internal/configuration"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

var APIPath string = "/api/inventory/v1"

func GetHost(insightsClientID string) (Host, error) {
	config := configuration.GetConfiguration()

	client, err := api.NewAuthenticatedClient(config.IdentityCertificate, config.IdentityKey, config.CACertificate)
	if err != nil {
		return Host{}, err
	}

	params := url.Values{}
	params.Set("insights_id", insightsClientID)
	endpoint := "https://" + config.APIHost + fmt.Sprintf(":%d", config.APIPort) + APIPath + "/hosts?" + params.Encode()

	req, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		slog.Error("could not construct request", slog.String("error", err.Error()))
		return Host{}, err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("could not contact HBI", slog.String("error", err.Error()))
		return Host{}, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("could not read response", slog.String("error", err.Error()))
		return Host{}, err
	}

	var hosts Hosts
	if err = json.Unmarshal(response, &hosts); err != nil {
		slog.Error("could not unmarshal response", slog.String("error", err.Error()))
		return Host{}, err
	}
	if len(hosts.Results) == 0 {
		slog.Debug("HBI returned no hosts", slog.String("response", string(response)))
		return Host{}, errors.New("no hosts found")
	}
	if len(hosts.Results) > 1 {
		slog.Debug("HBI returned more hosts than expected", slog.Int("count", len(hosts.Results)))
	}

	return hosts.Results[0], nil
}
