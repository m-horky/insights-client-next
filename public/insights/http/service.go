package http

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/m-horky/insights-client-next/internal/app"
)

// Service is a representation of an API service.
//
// To create an instance, use NewService.
type Service struct {
	Protocol string
	Hostname string
	Port     uint64
	// Path is in a form of `api/v1/NAME`, without leading or trailing slashes.
	Path string
}

func (s *Service) String() string {
	return fmt.Sprintf("%s://%s:%d/%s", s.Protocol, s.Hostname, s.Port, s.Path)
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

	fullUrl := fmt.Sprintf("%s/%s?%s", s, endpoint, parameters.Encode())

	if body == nil {
		body = bytes.NewBuffer(nil)
	}
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

	slog.Debug("request sent", slog.String("url", fullUrl), slog.Any("headers", req.Header))

	if os.Getenv("HTTP_DEBUG") != "" && body.Len() > 0 {
		slog.Debug("request data", slog.String("payload", body.String()))
	}

	now := time.Now()
	resp, err := client.Do(req)
	delta := time.Since(now)
	if err != nil {
		slog.Error("could not make request", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()
	slog.Debug(
		"response received",
		slog.Int("code", resp.StatusCode),
		slog.Duration("rtt", delta),
	)

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("could not read response body", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	if os.Getenv("HTTP_DEBUG") != "" && len(response) > 0 {
		slog.Debug("response data", slog.String("payload", string(response)))
	}

	return &Response{Code: resp.StatusCode, Data: response}, nil
}
