package oidc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rancher/shepherd/pkg/clientbase"
)

const (
	OIDCClientGroup    = "management.cattle.io"
	OIDCClientVersion  = "v3"
	OIDCClientResource = "oidcclients"
	OIDCClientKind     = "OIDCClient"

	OIDCDiscoveryPath = "/oidc/.well-known/openid-configuration"
	OIDCAuthPath      = "/oidc/authorize"
	OIDCTokenPath     = "/oidc/token"
	UsersPath         = "/v3/users"
	ClustersPath      = "/v3/clusters"
)

type TokenSet struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// Discovery fetches the OIDC provider metadata document.
func Discovery(httpClient *http.Client, rancherURL string) (*clientbase.Response, map[string]interface{}, error) {
	resp, err := clientbase.Do(httpClient, "GET", rancherURL+OIDCDiscoveryPath, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching OIDC discovery: %w", err)
	}
	var doc map[string]interface{}
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		return resp, nil, fmt.Errorf("parsing discovery document: %w", err)
	}
	return resp, doc, nil
}

// RefreshAccessToken exchanges a refresh_token for a new TokenSet.
func RefreshAccessToken(httpClient *http.Client, rancherURL, refreshToken, clientID, clientSecret string) (*TokenSet, error) {
	body := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	resp, err := clientbase.Do(httpClient, "POST", rancherURL+OIDCTokenPath, body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return nil, fmt.Errorf("refresh token POST: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh returned %d: %s", resp.StatusCode, resp.Body)
	}
	var ts TokenSet
	if err := json.Unmarshal(resp.Body, &ts); err != nil {
		return nil, fmt.Errorf("parsing refresh response: %w", err)
	}
	return &ts, nil
}
