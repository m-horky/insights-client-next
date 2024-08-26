package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/m-horky/insights-client-next/internal/app"
)

// Service is a representation of an API service.
//
// To create an instance, use NewService.
type Service struct {
	apiPath string
}

// NewService creates an instance of Service.
//
// When making a connection, the API path is appended to the URL.
// It should be in a form of `api/v1/NAME`, without leading or trailing slashes.
func NewService(apiPath string) Service {
	return Service{apiPath: apiPath}
}

// MakeRequest sends a request to a relevant service.
//
// This method uses RHSM certificates to authenticate to the server.
// Unless present, the `Accept` header is set to `application/json`.
func (s *Service) MakeRequest(
	method,
	endpoint string,
	parameters url.Values,
	headers map[string][]string,
	body *bytes.Buffer,
) (*Response, error) {
	config := app.GetConfiguration()

	fullUrl := fmt.Sprintf(
		"%s://%s:%d/%s/%s?%s",
		config.APIProtocol,
		config.APIHost,
		config.APIPort,
		s.apiPath,
		endpoint,
		parameters.Encode(),
	)

	req, err := http.NewRequest(method, fullUrl, body)
	if err != nil {
		slog.Error("could not construct request", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not construct request: %w", err)
	}

	for key, value := range headers {
		req.Header[key] = value
	}
	if _, ok := req.Header["Accept"]; !ok {
		// Default to requesting JSON
		req.Header.Set("Accept", "application/json")
	}

	client, err := NewAuthenticatedClient(config.IdentityCertificate, config.IdentityKey)
	if err != nil {
		slog.Error("could not create client", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	slog.Debug("making request", slog.String("url", fullUrl), slog.Any("headers", req.Header))
	now := time.Now()
	resp, err := client.Do(req)
	delta := time.Since(now)
	if err != nil {
		slog.Error("could not make request", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()
	slog.Debug("response received", slog.Int("code", resp.StatusCode), slog.Duration("rtt", delta))

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("could not read response body", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	return &Response{Code: resp.StatusCode, Data: response}, nil
}
