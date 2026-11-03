package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/clientbase"
)

const (
	logoutPath      = "/v1/logout"
	logoutAllAction = "?action=logoutAll"
	managementPath  = "/v3"
)

// LogoutAll ends the session at the identity provider and returns the address the browser is sent to
func LogoutAll(client *rancher.Client, finalRedirectURL string) (string, error) {
	response, err := postLogout(client, logoutPath+logoutAllAction, map[string]any{
		"finalRedirectUrl": finalRedirectURL,
	})
	if err != nil {
		return "", err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Rancher refused the logout-all, it answered %d: %s", response.StatusCode,
			response.Body)
	}

	var output struct {
		IDPRedirectURL string `json:"idpRedirectUrl"`
	}

	if err := json.Unmarshal(response.Body, &output); err != nil {
		return "", fmt.Errorf("decoding the logout-all answer: %w", err)
	}

	if output.IDPRedirectURL == "" {
		return "", fmt.Errorf("Rancher accepted the logout-all but named no address to end the identity "+
			"provider session at: %s", response.Body)
	}

	return output.IDPRedirectURL, nil
}

// Logout ends the current Rancher session
func Logout(client *rancher.Client) error {
	response, err := postLogout(client, logoutPath, map[string]any{})
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Rancher refused the logout, it answered %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

func postLogout(client *rancher.Client, path string, body map[string]any) (*clientbase.Response, error) {
	rancherURL := strings.TrimSuffix(client.Management.APIBaseClient.Opts.URL, managementPath)

	insecure := false
	if client.RancherConfig.Insecure != nil {
		insecure = *client.RancherConfig.Insecure
	}

	httpClient := &http.Client{
		Timeout: defaults.OneMinuteTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, // nolint:gosec
		},
	}

	response, err := clientbase.Do(httpClient, http.MethodPost, rancherURL+path, body, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + client.Management.APIBaseClient.Opts.TokenKey,
	})
	if err != nil {
		return nil, fmt.Errorf("posting %s: %w", path, err)
	}

	return response, nil
}
