package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func NewAuthenticatedClient(certPath, keyPath string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		slog.Error("could not load identity certificate", slog.Any("error", err))
		return nil, fmt.Errorf("could not load identity certificate: %w", err)
	}

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		slog.Error("could not read CA certificate", slog.Any("error", err))
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		slog.Error("could not load system cert pool", slog.Any("error", err))
		return nil, fmt.Errorf("could not load system cert pool: %w", err)
	}
	pool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}, nil
}
