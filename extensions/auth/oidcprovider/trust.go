package oidcprovider

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/rancher/shepherd/extensions/defaults"
)

const (
	certificatePEMType = "CERTIFICATE"
	defaultTLSPort     = "443"
)

// FetchServerCertificates returns the PEM encoded certificates an address serves for TLS
func FetchServerCertificates(address string) (string, error) {
	dialer := &net.Dialer{Timeout: defaults.OneMinuteTimeout}

	connection, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{InsecureSkipVerify: true}) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("opening a TLS connection to %s to read the certificate it serves: %w", address, err)
	}
	defer connection.Close()

	presented := connection.ConnectionState().PeerCertificates
	if len(presented) == 0 {
		return "", fmt.Errorf("%s completed a TLS handshake without presenting a certificate", address)
	}

	var encoded strings.Builder

	for _, certificate := range presented {
		if err := pem.Encode(&encoded, &pem.Block{Type: certificatePEMType, Bytes: certificate.Raw}); err != nil {
			return "", fmt.Errorf("encoding the certificate %s serves: %w", address, err)
		}
	}

	return encoded.String(), nil
}

// TLSAddressOf returns the host and port a URL is reached over TLS at
func TLSAddressOf(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing %s to find the address it is served from: %w", rawURL, err)
	}

	if parsed.Hostname() == "" {
		return "", fmt.Errorf("%s names no host, so there is no address to read a certificate from", rawURL)
	}

	port := parsed.Port()
	if port == "" {
		port = defaultTLSPort
	}

	return net.JoinHostPort(parsed.Hostname(), port), nil
}
