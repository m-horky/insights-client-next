package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/m-horky/insights-client-next/internal/configuration"
)

type Service struct {
	apiPath string
}

func NewService(apiPath string) Service {
	return Service{apiPath: apiPath}
}

func (s *Service) MakeRequest(
	method,
	endpoint string,
	parameters url.Values,
	headers map[string][]string,
	body *bytes.Buffer,
) (*Response, error) {
	config := configuration.GetConfiguration()

	fullUrl := fmt.Sprintf("%s://%s:%d/%s/%s?%s",
		config.APIProtocol, config.APIHost, config.APIPort,
		s.apiPath, endpoint, parameters.Encode(),
	)

	req, err := http.NewRequest(method, fullUrl, body)
	if err != nil {
		slog.Error("could not construct request", slog.Any("error", err))
		return nil, fmt.Errorf("could not construct request: %w", err)
	}

	for key, value := range headers {
		req.Header[key] = value
	}
	if _, ok := req.Header["Accept"]; !ok {
		// FIXME I guess we should be defaulting to JSON, right?
		req.Header.Set("Accept", "application/json")
	}

	client, err := NewAuthenticatedClient(config.IdentityCertificate, config.IdentityKey)
	if err != nil {
		slog.Error("could not create client", slog.Any("error", err))
		return nil, fmt.Errorf("could not create client: %w", err)
	}

	slog.Debug(
		"making request",
		slog.String("url", fullUrl),
		slog.Any("headers", req.Header),
	)
	now := time.Now()
	resp, err := client.Do(req)
	delta := time.Since(now)
	if err != nil {
		slog.Error("could not send request", slog.Any("error", err))
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	slog.Debug(
		"response received",
		slog.Int("code", resp.StatusCode),
		slog.Duration("rtt", delta),
	)
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("could not read response", slog.Any("error", err))
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	// TODO Is it smart to load everything into memory?
	return &Response{Code: resp.StatusCode, Data: response}, nil
}
