package saml

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ServiceProviderMetadata represents the service provider descriptor Rancher publishes for a SAML provider.
type ServiceProviderMetadata struct {
	EntityID     string
	Certificates []string
	Raw          string
}

// HasCertificate reports whether the descriptor carries the given certificate
func (m *ServiceProviderMetadata) HasCertificate(certificate string) bool {
	wanted := normalizeCertificate(certificate)
	if wanted == "" {
		return false
	}

	for _, held := range m.Certificates {
		if normalizeCertificate(held) == wanted {
			return true
		}
	}

	return false
}

// FetchServiceProviderMetadata reads the service provider descriptor Rancher publishes for a provider
func FetchServiceProviderMetadata(httpClient *http.Client, rancherURL, hostHeader, provider string) (*ServiceProviderMetadata, error) {
	requestURL := strings.TrimRight(rancherURL, "/") + MetadataPath(provider)

	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building a service provider metadata request to %s: %w", requestURL, err)
	}
	if hostHeader != "" {
		request.Host = hostHeader
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting the service provider metadata from %s: %w", requestURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the service provider metadata from %s: %w", requestURL, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %d rather than the service provider metadata: %s",
			requestURL, response.StatusCode, string(body))
	}

	descriptor := &struct {
		EntityID     string   `xml:"entityID,attr"`
		Certificates []string `xml:"SPSSODescriptor>KeyDescriptor>KeyInfo>X509Data>X509Certificate"`
	}{}
	if err := xml.Unmarshal(body, descriptor); err != nil {
		return nil, fmt.Errorf("decoding the service provider metadata from %s, %s: %w",
			requestURL, string(body), err)
	}

	return &ServiceProviderMetadata{
		EntityID:     descriptor.EntityID,
		Certificates: descriptor.Certificates,
		Raw:          string(body),
	}, nil
}

func normalizeCertificate(certificate string) string {
	replacer := strings.NewReplacer(
		"-----BEGIN CERTIFICATE-----", "",
		"-----END CERTIFICATE-----", "",
		"\r", "",
		"\n", "",
		"\t", "",
		" ", "",
	)

	return replacer.Replace(certificate)
}
