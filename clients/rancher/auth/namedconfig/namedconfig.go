package namedconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/pkg/clientbase"
	"github.com/rancher/shepherd/pkg/session"
)

const (
	configNameField         = "configName"
	managementAPIPath       = "/v3"
	providerTypesPath       = "/v1-public/authprovider-types"
	defaultNewConfigAccess  = "unrestricted"
	jsonContentType         = "application/json"
	contentTypeHeader       = "Content-Type"
	acceptHeader            = "Accept"
	unexpectedProviderTypes = "unexpected response listing the auth provider types"
	testAndApplyAction      = "testAndApply"
	disableAction           = "disable"
)

// NamedAuthConfig identifies one auth config of a provider. An empty ConfigName addresses the provider's default config.
type NamedAuthConfig struct {
	SchemaType   string
	ResourceType string
	ConfigType   string
	ConfigName   string
	DisplayName  string
	AccessMode   string
}

func (t NamedAuthConfig) name() string {
	if t.ConfigName != "" {
		return t.ConfigName
	}

	return t.ResourceType
}

// Enable creates the named config when it does not exist yet, applies input through the testAndApply action, and registers cleanup.
// The config is disabled on cleanup, and deleted too when this call created it.
func Enable(sess *session.Session, client *management.Client, config NamedAuthConfig, input any) error {
	body, err := withConfigName(input, config.ConfigName)
	if err != nil {
		return err
	}

	created, err := ensure(client, config)
	if err != nil {
		return err
	}

	if created {
		// Registered before the disable cleanup below, so the session disables the config first and then deletes it.
		sess.RegisterCleanupFunc(func() error {
			return Delete(client, config)
		})
	}

	var jsonResp map[string]any
	if err := client.Ops.DoModify("POST", actionURL(client, config, testAndApplyAction), body, &jsonResp); err != nil {
		return err
	}

	sess.RegisterCleanupFunc(func() error {
		return Disable(client, config)
	})

	return nil
}

// Disable turns the config off. A named config that is already disabled or gone is left alone, since Rancher rejects the action on it.
func Disable(client *management.Client, config NamedAuthConfig) error {
	if config.ConfigName != "" {
		enabled, err := isEnabled(client, config.ConfigName)
		if err != nil {
			return err
		}

		if !enabled {
			return nil
		}
	}

	var jsonResp map[string]any
	disableInput := map[string]string{configNameField: config.name()}

	return client.Ops.DoModify("POST", actionURL(client, config, disableAction), disableInput, &jsonResp)
}

// Delete removes the named config, treating one that is already gone as deleted. The provider's default config cannot be deleted.
func Delete(client *management.Client, config NamedAuthConfig) error {
	if config.ConfigName == "" {
		return fmt.Errorf("the default %s auth config cannot be deleted", config.ResourceType)
	}

	existing, err := client.AuthConfig.ByID(config.ConfigName)
	if err != nil {
		if clientbase.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("retrieving the %s auth config: %w", config.ConfigName, err)
	}

	if err := client.AuthConfig.Delete(existing); err != nil && !clientbase.IsNotFound(err) {
		return fmt.Errorf("deleting the %s auth config: %w", config.ConfigName, err)
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

	var collection map[string]map[string]any
	if err := json.Unmarshal(response.Body, &collection); err != nil {
		return nil, fmt.Errorf("%s: %w (body: %s)", unexpectedProviderTypes, err, response.Body)
	}

	publicTypes := make([]string, 0, len(collection))
	for publicType := range collection {
		publicTypes = append(publicTypes, publicType)
	}
	sort.Strings(publicTypes)

	providerTypes := make([]map[string]any, 0, len(publicTypes))
	for _, publicType := range publicTypes {
		entry := collection[publicType]
		entry["providerType"] = publicType
		providerTypes = append(providerTypes, entry)
	}

	return providerTypes, nil
}

// withConfigName returns input as a request body that names the config it applies to. An empty configName returns input untouched.
func withConfigName(input any, configName string) (any, error) {
	if configName == "" {
		return input, nil
	}

	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encoding the request body: %w", err)
	}

	body := map[string]any{}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("adding %s to the request body: %w", configNameField, err)
	}

	body[configNameField] = configName

	return body, nil
}

// ensure creates the named config when it does not exist yet and reports whether it did.
func ensure(client *management.Client, config NamedAuthConfig) (bool, error) {
	if config.ConfigName == "" {
		return false, nil
	}

	_, err := client.AuthConfig.ByID(config.ConfigName)
	if err == nil {
		return false, nil
	}

	if !clientbase.IsNotFound(err) {
		return false, fmt.Errorf("retrieving the %s auth config: %w", config.ConfigName, err)
	}

	accessMode := config.AccessMode
	if accessMode == "" {
		accessMode = defaultNewConfigAccess
	}

	_, err = client.AuthConfig.Create(&management.AuthConfig{
		Name:        config.ConfigName,
		Type:        config.ConfigType,
		DisplayName: config.DisplayName,
		AccessMode:  accessMode,
	})
	if err != nil {
		return false, fmt.Errorf("creating the %s auth config: %w", config.ConfigName, err)
	}

	return true, nil
}

// isEnabled reports whether the named config exists and is enabled. A config that is gone counts as not enabled.
func isEnabled(client *management.Client, configName string) (bool, error) {
	existing, err := client.AuthConfig.ByID(configName)
	if err != nil {
		if clientbase.IsNotFound(err) {
			return false, nil
		}

		return false, fmt.Errorf("retrieving the %s auth config: %w", configName, err)
	}

	return existing.Enabled, nil
}

// actionURL returns the URL of an action on the config the config addresses.
func actionURL(client *management.Client, config NamedAuthConfig, action string) string {
	return fmt.Sprintf("%v/%v/%v?action=%v", client.Opts.URL, config.SchemaType, config.name(), action)
}
