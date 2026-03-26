package scim

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rancher/shepherd/pkg/clientbase"
)

const (
	SCIMSchemaUser    = "urn:ietf:params:scim:schemas:core:2.0:User"
	SCIMSchemaGroup   = "urn:ietf:params:scim:schemas:core:2.0:Group"
	SCIMSchemaPatchOp = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

type User struct {
	Schemas    []string `json:"schemas"`
	UserName   string   `json:"userName"`
	ExternalID string   `json:"externalId,omitempty"`
	Active     *bool    `json:"active,omitempty"`
}

type Group struct {
	Schemas     []string `json:"schemas"`
	ID          string   `json:"id,omitempty"`
	DisplayName string   `json:"displayName"`
	ExternalID  string   `json:"externalId,omitempty"`
	Members     []Member `json:"members,omitempty"`
}

type Member struct {
	Value string `json:"value"`
}

type PatchOp struct {
	Schemas    []string    `json:"schemas"`
	Operations []Operation `json:"Operations"`
}

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

func (r *Response) DecodeJSON(target interface{}) error {
	return json.Unmarshal(r.Body, target)
}

func (r *Response) IDFromBody() (string, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(r.Body, &m); err != nil {
		return "", err
	}
	id, ok := m["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("id not found in response: %s", string(r.Body))
	}
	return id, nil
}

func BoolPtr(b bool) *bool { return &b }

type scimTransport struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func (t *scimTransport) do(method, resource, id string, query url.Values, body interface{}) (*Response, error) {
	rawURL := fmt.Sprintf("%s/%s", t.baseURL, resource)
	if id != "" {
		rawURL += "/" + id
	}
	if len(query) > 0 {
		rawURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &Response{StatusCode: resp.StatusCode, Body: respBody, Header: resp.Header}, nil
}

// Users is the interface for SCIM User operations.
type Users interface {
	List(query url.Values) (*Response, error)
	Create(user User) (*Response, error)
	ByID(id string) (*Response, error)
	Update(id string, user User) (*Response, error)
	Patch(id string, patch PatchOp) (*Response, error)
	Delete(id string) (*Response, error)
}

// Groups is the interface for SCIM Group operations.
type Groups interface {
	List(query url.Values) (*Response, error)
	Create(group Group) (*Response, error)
	ByID(id string) (*Response, error)
	ByIDWithQuery(id string, query url.Values) (*Response, error)
	Update(id string, group Group) (*Response, error)
	Patch(id string, patch PatchOp) (*Response, error)
	Delete(id string) (*Response, error)
}

// Discovery is the interface for SCIM discovery operations.
type Discovery interface {
	ServiceProviderConfig() (*Response, error)
	ResourceTypes() (*Response, error)
	ResourceTypeByID(id string) (*Response, error)
	Schemas() (*Response, error)
	SchemaByID(id string) (*Response, error)
}

type Client struct {
	t *scimTransport
}

// Build the SCIM base URL by replacing any existing path on opts.URL.
// opts.URL may carry an API prefix (e.g. "/v3") so we parse and reset
// the path rather than appending to avoid paths like "/v3/v1-scim/...".

func NewClient(opts *clientbase.ClientOpts, provider string) *Client {
	baseURL := fmt.Sprintf("%s/v1-scim/%s", opts.URL, provider)
	if u, err := url.Parse(opts.URL); err == nil {
		u.Path = fmt.Sprintf("/v1-scim/%s", provider)
		u.RawQuery = ""
		u.Fragment = ""
		baseURL = u.String()
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		if opts.Insecure {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else if opts.CACerts != "" {
			roots := x509.NewCertPool()
			roots.AppendCertsFromPEM([]byte(opts.CACerts))
			tr.TLSClientConfig = &tls.Config{RootCAs: roots}
		}
		timeout := opts.Timeout
		if timeout == 0 {
			timeout = time.Minute
		}
		httpClient = &http.Client{Transport: tr, Timeout: timeout}
	}

	return &Client{t: &scimTransport{
		baseURL:    baseURL,
		token:      opts.TokenKey,
		httpClient: httpClient,
	}}
}

func (c *Client) Users() Users {
	return &userClient{t: c.t}
}

func (c *Client) Groups() Groups {
	return &groupClient{t: c.t}
}

func (c *Client) Discovery() Discovery {
	return &discoveryClient{t: c.t}
}

type userClient struct {
	t *scimTransport
}

func (c *userClient) List(query url.Values) (*Response, error) {
	return c.t.do(http.MethodGet, "Users", "", query, nil)
}

func (c *userClient) Create(user User) (*Response, error) {
	return c.t.do(http.MethodPost, "Users", "", nil, user)
}

func (c *userClient) ByID(id string) (*Response, error) {
	return c.t.do(http.MethodGet, "Users", id, nil, nil)
}

func (c *userClient) Update(id string, user User) (*Response, error) {
	return c.t.do(http.MethodPut, "Users", id, nil, user)
}

func (c *userClient) Patch(id string, patch PatchOp) (*Response, error) {
	return c.t.do(http.MethodPatch, "Users", id, nil, patch)
}

func (c *userClient) Delete(id string) (*Response, error) {
	return c.t.do(http.MethodDelete, "Users", id, nil, nil)
}

type groupClient struct {
	t *scimTransport
}

func (c *groupClient) List(query url.Values) (*Response, error) {
	return c.t.do(http.MethodGet, "Groups", "", query, nil)
}

func (c *groupClient) Create(group Group) (*Response, error) {
	return c.t.do(http.MethodPost, "Groups", "", nil, group)
}

func (c *groupClient) ByID(id string) (*Response, error) {
	return c.t.do(http.MethodGet, "Groups", id, nil, nil)
}

func (c *groupClient) ByIDWithQuery(id string, query url.Values) (*Response, error) {
	return c.t.do(http.MethodGet, "Groups", id, query, nil)
}

func (c *groupClient) Update(id string, group Group) (*Response, error) {
	return c.t.do(http.MethodPut, "Groups", id, nil, group)
}

func (c *groupClient) Patch(id string, patch PatchOp) (*Response, error) {
	return c.t.do(http.MethodPatch, "Groups", id, nil, patch)
}

func (c *groupClient) Delete(id string) (*Response, error) {
	return c.t.do(http.MethodDelete, "Groups", id, nil, nil)
}

type discoveryClient struct {
	t *scimTransport
}

func (c *discoveryClient) ServiceProviderConfig() (*Response, error) {
	return c.t.do(http.MethodGet, "ServiceProviderConfig", "", nil, nil)
}

func (c *discoveryClient) ResourceTypes() (*Response, error) {
	return c.t.do(http.MethodGet, "ResourceTypes", "", nil, nil)
}

func (c *discoveryClient) ResourceTypeByID(id string) (*Response, error) {
	return c.t.do(http.MethodGet, "ResourceTypes", id, nil, nil)
}

func (c *discoveryClient) Schemas() (*Response, error) {
	return c.t.do(http.MethodGet, "Schemas", "", nil, nil)
}

func (c *discoveryClient) SchemaByID(id string) (*Response, error) {
	return c.t.do(http.MethodGet, "Schemas", id, nil, nil)
}
