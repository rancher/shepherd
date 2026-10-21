package metaproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/clientbase"
)

const (
	ProxyPrefix = "/meta/proxy"

	AuthorizationHeader = "Authorization"
	CookieHeader        = "Cookie"
	APIAuthHeader       = "X-API-Auth-Header"
	CattleAuthHeader    = "X-API-CattleAuth-Header"
	APICookieHeader     = "X-Api-Cookie-Header"
	APISetCookieHeader  = "X-Api-Set-Cookie-Header"

	BearerSigner    = "bearer"
	BasicSigner     = "basic"
	DigestSigner    = "digest"
	AWSV4Signer     = "awsv4"
	ArbitrarySigner = "arbitrary"
)

// Request sends a request through the Rancher meta proxy to the given destination carrying exactly the given header
func Request(client *rancher.Client, method, destination string, body interface{}, header http.Header) (*clientbase.Response, error) {
	proxyURL, err := URL(client, destination)
	if err != nil {
		return nil, err
	}

	return clientbase.DoWithHeader(client.Management.APIBaseClient.Ops.Client, method, proxyURL, body, header)
}

// URL returns the Rancher meta proxy URL that forwards to the given destination
func URL(client *rancher.Client, destination string) (string, error) {
	if client.RancherConfig == nil || client.RancherConfig.Host == "" {
		return "", fmt.Errorf("the rancher host is not set on the client")
	}

	destinationPath, rawQuery, _ := strings.Cut(destination, "?")

	proxyURL := &url.URL{
		Scheme:   "https",
		Host:     client.RancherConfig.Host,
		Path:     path.Join(ProxyPrefix, destinationPath),
		RawQuery: rawQuery,
	}

	if strings.HasSuffix(destinationPath, "/") && !strings.HasSuffix(proxyURL.Path, "/") {
		proxyURL.Path += "/"
	}

	return proxyURL.String(), nil
}
