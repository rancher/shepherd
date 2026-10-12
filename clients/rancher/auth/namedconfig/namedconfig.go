package namedconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/clientbase"
)

const (
	// ConfigNameField is the request body field that names the auth config an action applies to
	ConfigNameField = "configName"

	managementAPIPath       = "/v3"
	providerTypesPath       = "/v1-public/authprovider-types"
	defaultNewConfigAccess  = "unrestricted"
	collectionDataField     = "data"
	jsonContentType         = "application/json"
	contentTypeHeader       = "Content-Type"
	acceptHeader            = "Accept"
	unexpectedProviderTypes = "unexpected response listing the auth provider types"
)

// Body returns input as a request body that names the auth config it applies to.
// An empty configName returns input untouched, so the provider's default config is used.
func Body(input any, configName string) (any, error) {
	if configName == "" {
		return input, nil
	}

	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encoding the request body: %w", err)
	}

	body := map[string]any{}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("adding %s to the request body: %w", ConfigNameField, err)
	}

	body[ConfigNameField] = configName

	return body, nil
}

// DisableBody returns the body of the disable action for the named auth config
func DisableBody(configName string) map[string]any {
	body := map[string]any{"action": "disable"}
	if configName != "" {
		body[ConfigNameField] = configName
	}

	return body
}

// Ensure creates the named auth config when it does not exist yet. Created configs are deleted when the client's session is cleaned up.
func Ensure(client *management.Client, configType, configName, displayName string) error {
	if configName == "" {
		return nil
	}

	_, err := client.AuthConfig.ByID(configName)
	if err == nil {
		return nil
	}

	if !clientbase.IsNotFound(err) {
		return fmt.Errorf("retrieving the %s auth config: %w", configName, err)
	}

	_, err = client.AuthConfig.Create(&management.AuthConfig{
		Name:        configName,
		Type:        configType,
		DisplayName: displayName,
		AccessMode:  defaultNewConfigAccess,
	})
	if err != nil {
		return fmt.Errorf("creating the %s auth config: %w", configName, err)
	}

	return nil
}

// Delete removes the named auth config, treating one that is already gone as deleted
func Delete(client *management.Client, configName string) error {
	existing, err := client.AuthConfig.ByID(configName)
	if err != nil {
		if clientbase.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("retrieving the %s auth config: %w", configName, err)
	}

	if err := client.AuthConfig.Delete(existing); err != nil && !clientbase.IsNotFound(err) {
		return fmt.Errorf("deleting the %s auth config: %w", configName, err)
	}

	return nil
}

// ListProviderTypes returns what GET /v1-public/authprovider-types reports about the auth providers Rancher offers
func ListProviderTypes(client *management.Client) ([]map[string]any, error) {
	rancherURL := strings.TrimSuffix(client.Opts.URL, managementAPIPath)

	response, err := client.Ops.RawHttpRequest(http.MethodGet, rancherURL+providerTypesPath, nil,
		map[string]string{acceptHeader: jsonContentType, contentTypeHeader: jsonContentType})
	if err != nil {
		return nil, fmt.Errorf("listing the auth provider types: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("listing the auth provider types, Rancher answered %d: %s", response.StatusCode, response.Body)
	}

	var providerTypes []map[string]any
	if err := json.Unmarshal(response.Body, &providerTypes); err == nil {
		return providerTypes, nil
	}

	var collection map[string]json.RawMessage
	if err := json.Unmarshal(response.Body, &collection); err != nil {
		return nil, fmt.Errorf("%s: %s", unexpectedProviderTypes, response.Body)
	}

	data, ok := collection[collectionDataField]
	if !ok {
		return nil, fmt.Errorf("%s: %s", unexpectedProviderTypes, response.Body)
	}

	if err := json.Unmarshal(data, &providerTypes); err != nil {
		return nil, fmt.Errorf("%s: %s", unexpectedProviderTypes, response.Body)
	}

	return providerTypes, nil
}
