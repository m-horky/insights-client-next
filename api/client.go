package api

import (
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

// NewAuthenticatedClient creates a client that uses mTLS authentication.
func NewAuthenticatedClient(certPath, keyPath string, proxy *url.URL) (*http.Client, *Error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		slog.Error("could not load identity certificate", slog.String("error", err.Error()))
		return nil, NewError(ErrNoCertificate, err, nil, "Could not load identity certificate.")
	}

	caCert, err := os.ReadFile(certPath)
	if err != nil {
		slog.Error("could not load CA certificate", slog.String("error", err.Error()))
		return nil, NewError(ErrNoCertificate, err, nil, "Could not load system certificates.")
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		slog.Error("could not load system certificate pool", slog.String("error", err.Error()))
		return nil, NewError(ErrNoCertificate, err, nil, "Could not load system certificates.")
	}

	pool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{Transport: transport}, nil
}
