package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/m-horky/insights-client-next/app"
)

type ServiceURL struct {
	protocol string
	hostname string
	port     uint
}

// NewServiceURL creates a definition for service API.
func NewServiceURL(protocol, hostname string, port uint) *ServiceURL {
	return &ServiceURL{
		protocol: protocol,
		hostname: hostname,
		port:     port,
	}
}

// Service is a representation of an API service.
//
// To create an instance, use NewService.
type Service struct {
	URL               *ServiceURL
	Path              string
	ClientCertificate string
	ClientKey         string
}

func NewService(url *ServiceURL) *Service {
	return &Service{URL: url}
}

func NewServiceWithAuthentication(url *ServiceURL, certificate, key string) *Service {
	return &Service{URL: url, ClientCertificate: certificate, ClientKey: key}
}

// String formats the service into a URI.
func (s *Service) String() string {
	return fmt.Sprintf("%s://%s:%d/%s", s.URL.protocol, s.URL.hostname, s.URL.port, s.Path)
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
) (*Response, app.HumanError) {
	fullUrl := fmt.Sprintf("%s/%s?%s", s, endpoint, parameters.Encode())

	if body == nil {
		body = bytes.NewBuffer(nil)
	}
	req, err := http.NewRequest(method, fullUrl, body)
	if err != nil {
		slog.Error("could not construct request", slog.String("error", err.Error()))
		return nil, app.NewError(ErrRequest, err, "Could not construct API request.")
	}

	for key, value := range headers {
		req.Header[key] = value
	}
	if _, ok := req.Header["Accept"]; !ok {
		// Default to requesting JSON
		req.Header.Set("Accept", "application/json")
	}

	client, err := NewAuthenticatedClient(s.ClientCertificate, s.ClientKey)
	if err != nil {
		slog.Error("could not create client", slog.String("error", err.Error()))
		return nil, app.NewError(ErrRequest, err, "Could not create API client.")
	}

	slog.Debug("request sent", slog.String("URL", fullUrl), slog.Any("headers", req.Header))

	if os.Getenv("HTTP_DEBUG") != "" && body.Len() > 0 {
		slog.Debug("request data", slog.String("payload", stringifyData(body.Bytes())))
	}

	now := time.Now()
	resp, err := client.Do(req)
	delta := time.Since(now)
	if err != nil {
		slog.Error("could not make request", slog.String("error", err.Error()))
		return nil, app.NewError(ErrRequest, err, "Could not make API request.")
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
		return nil, app.NewError(ErrRequest, err, "Could not read API response.")
	}

	if os.Getenv("HTTP_DEBUG") != "" && len(response) > 0 {
		slog.Debug("response data", slog.String("payload", stringifyData(response)))
	}

	return &Response{Code: resp.StatusCode, Data: response}, nil
}

// stringifyData takes in a byte slice and converts it to string.
//
// Since the data may contain anything, bytes outside printable ASCII are
// converted to simplified representation.
func stringifyData(data []byte) string {
	result := make([]byte, len(data))

	unprintable := 0
	for _, char := range data {
		isPrintable := false
		if char == '\n' || char == '\r' || (char >= ' ' && char < 127) {
			isPrintable = true
		}

		if isPrintable && unprintable > 0 {
			result = append(result, '.')
		}
		if isPrintable {
			result = append(result, char)
			unprintable = 0
		} else {
			unprintable++
		}
	}

	return string(result)
}
