package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

// NewAuthenticatedClient creates a client that uses mTLS authentication.
func NewAuthenticatedClient(certPath, keyPath string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		slog.Error("could not load identity certificate", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not load identity certificate: %w", err)
	}

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		slog.Error("could not load CA certificate", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not load CA certificate: %w", err)
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		slog.Error("could not load system certificate pool", slog.String("error", err.Error()))
		return nil, fmt.Errorf("could not load system certificate pool: %w", err)
	}

	pool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}
