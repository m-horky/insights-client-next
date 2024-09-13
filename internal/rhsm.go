package internal

import (
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/m-horky/insights-client-next/app"
)

// ReadRHSMIdentity reads the CommonName from x.509 certificate.
//
// It loads the subscription-manager UUID.
func ReadRHSMIdentity(filename string) (string, app.HumanError) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", app.NewError(nil, err, "Could not load identity certificate.")
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return "", app.NewError(nil, err, "Could not load identity certificate.")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", app.NewError(nil, err, "Could not load identity certificate.")
	}

	return cert.Subject.CommonName, nil
}
