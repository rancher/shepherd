package saml

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	samlext "github.com/rancher/shepherd/extensions/auth/saml"
	"github.com/rancher/shepherd/extensions/defaults"
)

type APIClient struct {
	rancherURL       string
	rancherHost      string
	transport        *http.Transport
	noRedirectClient *http.Client
}

type UserSession struct {
	api       *APIClient
	username  string
	orgURL    string
	idpClient *http.Client
}

// NewAPIClient constructs an APIClient against the given Rancher URL
func NewAPIClient(rancherURL string) (*APIClient, error) {
	normalizedURL := strings.TrimRight(rancherURL, "/")
	if !strings.HasPrefix(normalizedURL, "https://") && !strings.HasPrefix(normalizedURL, "http://") {
		normalizedURL = "https://" + normalizedURL
	}

	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("parsing Rancher URL %s: %w", rancherURL, err)
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("Rancher URL %s has no host", rancherURL)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
	}

	return &APIClient{
		rancherURL:  normalizedURL,
		rancherHost: parsedURL.Host,
		transport:   transport,
		noRedirectClient: &http.Client{
			Transport: transport,
			Timeout:   defaults.OneMinuteTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// AuthenticateUser signs a user in at the identity provider and returns the resulting session
func (c *APIClient) AuthenticateUser(provider Provider, username, password string) (*UserSession, error) {
	loginRequest, err := samlext.InitiateLogin(c.noRedirectClient, c.rancherURL, provider.Name, provider.ConfigType,
		c.finalRedirectURL())
	if err != nil {
		return nil, fmt.Errorf("discovering the identity provider for %s: %w", provider.Name, err)
	}

	userSession, _, err := c.AuthenticateForRequest(provider, username, password, loginRequest)

	return userSession, err
}

// AuthenticateForRequest signs a user in against a single sign-on request Rancher has already made
func (c *APIClient) AuthenticateForRequest(provider Provider, username, password string,
	loginRequest *samlext.LoginRequest) (*UserSession, string, error) {
	idpURL, err := url.Parse(loginRequest.IdpRedirectURL)
	if err != nil {
		return nil, "", fmt.Errorf("parsing the identity provider redirect %s: %w", loginRequest.IdpRedirectURL, err)
	}
	orgURL := idpURL.Scheme + "://" + idpURL.Host

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating a cookie jar for the identity provider session: %w", err)
	}

	idpClient := &http.Client{
		Transport: c.transport,
		Timeout:   defaults.OneMinuteTimeout,
		Jar:       jar,
	}

	assertionDocument, err := samlext.AuthenticateWithLoginForm(idpClient, loginRequest.IdpRedirectURL, username, password)
	if err != nil {
		return nil, "", err
	}

	return &UserSession{
		api:       c,
		username:  username,
		orgURL:    orgURL,
		idpClient: idpClient,
	}, assertionDocument, nil
}

// AssertionForRequest returns the assertion answering a single sign-on request
func (s *UserSession) AssertionForRequest(provider Provider, loginRequest *samlext.LoginRequest,
	assertionDocument string) (*samlext.CapturedAssertion, error) {
	if assertionDocument != "" {
		return samlext.AssertionFromDocument(assertionDocument, provider.Name, loginRequest)
	}

	return samlext.FetchAssertion(s.idpClient, loginRequest, provider.Name)
}

// Login runs a complete service provider initiated SAML login and returns Rancher's answer
func (c *APIClient) Login(provider Provider, username, password string) (*samlext.ACSResult, error) {
	userSession, err := c.AuthenticateUser(provider, username, password)
	if err != nil {
		return nil, err
	}

	assertion, err := userSession.CaptureAssertion(provider)
	if err != nil {
		return nil, err
	}

	return c.SubmitAssertion(assertion, "")
}

// Username returns the identity provider user this session authenticated
func (s *UserSession) Username() string {
	return s.username
}

// CaptureAssertion returns an unused assertion without presenting it to Rancher
func (s *UserSession) CaptureAssertion(provider Provider) (*samlext.CapturedAssertion, error) {
	loginRequest, err := samlext.InitiateLogin(s.api.noRedirectClient, s.api.rancherURL, provider.Name, provider.ConfigType,
		s.api.finalRedirectURL())
	if err != nil {
		return nil, fmt.Errorf("initiating a %s login for %s: %w", provider.Name, s.username, err)
	}

	assertion, err := samlext.FetchAssertion(s.idpClient, loginRequest, provider.Name)
	if err != nil {
		return nil, fmt.Errorf("capturing an assertion for %s: %w", s.username, err)
	}

	return assertion, nil
}

// SubmitAssertion presents an assertion to Rancher's assertion consumer service
func (c *APIClient) SubmitAssertion(assertion *samlext.CapturedAssertion, targetBaseURL string) (*samlext.ACSResult, error) {
	if targetBaseURL == "" {
		targetBaseURL = c.rancherURL
	}

	return samlext.SubmitAssertion(c.noRedirectClient, targetBaseURL, c.rancherHost, assertion)
}

// FetchSessionIdentity resolves a Rancher session cookie to the user it authenticates as
func (c *APIClient) FetchSessionIdentity(sessionToken, targetBaseURL string) (*samlext.SessionIdentity, error) {
	if targetBaseURL == "" {
		targetBaseURL = c.rancherURL
	}

	return samlext.FetchSessionIdentity(c.noRedirectClient, targetBaseURL, c.rancherHost, sessionToken)
}

func (c *APIClient) finalRedirectURL() string {
	return c.rancherURL + samlext.DashboardVerifyPath
}
