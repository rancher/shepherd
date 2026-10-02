package oidcprovider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	oidcext "github.com/rancher/shepherd/extensions/auth/oidc"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/clientbase"
)

func exchangeAuthorizationCode(httpClient *http.Client, tokenEndpoint, clientID, clientSecret, redirectURI, code string) (*oidcext.TokenSet, error) {
	body := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	response, err := clientbase.Do(httpClient, http.MethodPost, tokenEndpoint, body, map[string]string{
		"Content-Type": formContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("exchanging an authorization code at %s: %w", tokenEndpoint, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the identity provider answered %d to the code exchange at %s: %s",
			response.StatusCode, tokenEndpoint, response.Body)
	}

	tokens := new(oidcext.TokenSet)
	if err := json.Unmarshal(response.Body, tokens); err != nil {
		return nil, fmt.Errorf("decoding the token response from %s: %w", tokenEndpoint, err)
	}

	if tokens.IDToken == "" {
		return nil, fmt.Errorf("the token response from %s carries no id_token, so no claim about the user can be read",
			tokenEndpoint)
	}

	return tokens, nil
}

// IDTokenClaims returns the claims an identity provider puts in the ID token it issues for a user
func IDTokenClaims(transport http.RoundTripper, authorizationURL, tokenEndpoint, clientID, clientSecret, redirectURI,
	username, password string) (map[string]interface{}, error) {
	code, err := CaptureAuthorizationCode(transport, authorizationURL, redirectURI, username, password)
	if err != nil {
		return nil, err
	}

	tokens, err := exchangeAuthorizationCode(&http.Client{Transport: transport, Timeout: defaults.OneMinuteTimeout},
		tokenEndpoint, clientID, clientSecret, redirectURI, code)
	if err != nil {
		return nil, err
	}

	return oidcext.DecodeJWTPayload(tokens.IDToken)
}

// ClaimStrings returns a claim's value as a list of strings, for claims an identity provider may send either way
func ClaimStrings(claims map[string]interface{}, name string) []string {
	switch value := claims[name].(type) {
	case string:
		return []string{value}
	case []interface{}:
		values := make([]string, 0, len(value))
		for _, entry := range value {
			if text, isText := entry.(string); isText {
				values = append(values, text)
			}
		}

		return values
	default:
		return nil
	}
}
