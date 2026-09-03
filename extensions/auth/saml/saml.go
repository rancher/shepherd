package saml

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	StateCookiePrefix   = "saml_"
	RancherStatePrefix  = "Rancher_"
	SessionCookieName   = "R_SESS"
	DashboardVerifyPath = "/dashboard/auth/verify"
	LoginAction         = "login"
	TestAndEnableAction = "testAndEnable"
	CookieResponseType  = "cookie"
	SAMLResponseField   = "SAMLResponse"
	RelayStateField     = "RelayState"

	acsPathFormat            = "/v1-saml/%s/saml/acs"
	metadataPathFormat       = "/v1-saml/%s/saml/metadata"
	currentUserPath          = "/v3/users?me=true"
	publicProviderPathFormat = "/v3-public/%sProviders/%s"
	configActionPathFormat   = "/v3/%ss/%s"
	configTypeSuffix         = "Config"
	loginDescription         = "shepherd saml login"
	formContentType          = "application/x-www-form-urlencoded"
	jsonContentType          = "application/json"
)

// ACSPath returns the assertion consumer service path Rancher serves for a provider
func ACSPath(provider string) string {
	return fmt.Sprintf(acsPathFormat, provider)
}

// MetadataPath returns the service provider metadata path Rancher serves for a provider
func MetadataPath(provider string) string {
	return fmt.Sprintf(metadataPathFormat, provider)
}

// PublicProviderPath returns the unauthenticated Norman path that exposes a provider's login action
func PublicProviderPath(provider, configType string) string {
	return fmt.Sprintf(publicProviderPathFormat, strings.TrimSuffix(configType, configTypeSuffix),
		strings.ToLower(provider))
}

// LoginRequest represents the state Rancher returns when a SAML login is initiated.
type LoginRequest struct {
	IdpRedirectURL string
	StateCookies   []*http.Cookie
	RelayState     string
}

// CapturedAssertion represents an assertion issued by the identity provider but not yet presented to Rancher.
type CapturedAssertion struct {
	Provider     string
	SAMLResponse string
	RelayState   string
	StateCookies []*http.Cookie
	Details      *AssertionDetails
}

// ACSResult represents the outcome of presenting an assertion to the assertion consumer service.
type ACSResult struct {
	Target       string
	StatusCode   int
	Location     string
	SessionToken string
	Accepted     bool
	Latency      time.Duration
}

// InitiateLogin starts a service provider initiated SAML login and returns the identity provider redirect
func InitiateLogin(httpClient *http.Client, rancherURL, provider, configType, finalRedirectURL string) (*LoginRequest, error) {
	payload, err := json.Marshal(map[string]string{
		"finalRedirectUrl": finalRedirectURL,
		"responseType":     CookieResponseType,
		"description":      loginDescription,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding the SAML login input: %w", err)
	}

	loginURL := strings.TrimRight(rancherURL, "/") + PublicProviderPath(provider, configType) + "?action=" + LoginAction
	request, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("building a login request to %s: %w", loginURL, err)
	}
	request.Header.Set("Content-Type", jsonContentType)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("posting the login action to %s: %w", loginURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the login action response from %s: %w", loginURL, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login action %s returned %d: %s", loginURL, response.StatusCode, string(body))
	}

	var loginOutput struct {
		IdpRedirectURL string `json:"idpRedirectUrl"`
	}
	if err := json.Unmarshal(body, &loginOutput); err != nil {
		return nil, fmt.Errorf("decoding the login action response %s: %w", string(body), err)
	}
	if loginOutput.IdpRedirectURL == "" {
		return nil, fmt.Errorf("login action %s returned no idpRedirectUrl: %s", loginURL, string(body))
	}

	return newLoginRequest(response, loginOutput.IdpRedirectURL, loginURL)
}

// InitiateTestAndEnable starts the round trip that turns a SAML provider on
func InitiateTestAndEnable(httpClient *http.Client, rancherURL, provider, configType, bearerToken, finalRedirectURL string) (*LoginRequest, error) {
	payload, err := json.Marshal(map[string]string{"finalRedirectUrl": finalRedirectURL})
	if err != nil {
		return nil, fmt.Errorf("encoding the SAML %s input: %w", TestAndEnableAction, err)
	}

	actionURL := strings.TrimRight(rancherURL, "/") +
		fmt.Sprintf(configActionPathFormat, configType, provider) + "?action=" + TestAndEnableAction

	request, err := http.NewRequest(http.MethodPost, actionURL, strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("building a %s request to %s: %w", TestAndEnableAction, actionURL, err)
	}
	request.Header.Set("Content-Type", jsonContentType)
	request.Header.Set("Authorization", "Bearer "+bearerToken)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("posting the %s action to %s: %w", TestAndEnableAction, actionURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the %s response from %s: %w", TestAndEnableAction, actionURL, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s action %s returned %d: %s",
			TestAndEnableAction, actionURL, response.StatusCode, string(body))
	}

	var actionOutput struct {
		IdpRedirectURL string `json:"idpRedirectUrl"`
	}
	if err := json.Unmarshal(body, &actionOutput); err != nil {
		return nil, fmt.Errorf("decoding the %s response %s: %w", TestAndEnableAction, string(body), err)
	}

	if actionOutput.IdpRedirectURL == "" {
		return nil, fmt.Errorf("%s action %s returned no idpRedirectUrl: %s",
			TestAndEnableAction, actionURL, string(body))
	}

	return newLoginRequest(response, actionOutput.IdpRedirectURL, actionURL)
}

// AssertionFromDocument reads an assertion out of a self-posting form an identity provider already served
func AssertionFromDocument(document, provider string, loginRequest *LoginRequest) (*CapturedAssertion, error) {
	form, err := ParseAutoSubmitForm(document)
	if err != nil {
		return nil, fmt.Errorf("the identity provider response holds no %s form: %w", SAMLResponseField, err)
	}

	samlResponse := form.Fields[SAMLResponseField]
	if samlResponse == "" {
		return nil, fmt.Errorf("the identity provider response holds a form posting to %s rather than an assertion",
			form.Action)
	}

	details, err := DecodeSAMLResponse(samlResponse)
	if err != nil {
		return nil, err
	}

	relayState := form.Fields[RelayStateField]
	if relayState == "" {
		relayState = loginRequest.RelayState
	}

	return &CapturedAssertion{
		Provider:     provider,
		SAMLResponse: samlResponse,
		RelayState:   relayState,
		StateCookies: loginRequest.StateCookies,
		Details:      details,
	}, nil
}

func newLoginRequest(response *http.Response, idpRedirectURL, source string) (*LoginRequest, error) {
	loginRequest := &LoginRequest{IdpRedirectURL: idpRedirectURL}

	for _, cookie := range response.Cookies() {
		if !strings.HasPrefix(cookie.Name, StateCookiePrefix) {
			continue
		}

		loginRequest.StateCookies = append(loginRequest.StateCookies, cookie)

		stateID := strings.TrimPrefix(cookie.Name, StateCookiePrefix)
		if !strings.HasPrefix(stateID, RancherStatePrefix) {
			loginRequest.RelayState = stateID
		}
	}

	if loginRequest.RelayState == "" {
		return nil, fmt.Errorf("%s set no %s request-state cookie", source, StateCookiePrefix)
	}

	return loginRequest, nil
}

// FetchAssertion follows an identity provider redirect and returns the assertion the provider issues
func FetchAssertion(idpClient *http.Client, loginRequest *LoginRequest, provider string) (*CapturedAssertion, error) {
	response, err := idpClient.Get(loginRequest.IdpRedirectURL)
	if err != nil {
		return nil, fmt.Errorf("following the identity provider redirect: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the identity provider response: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the identity provider returned %d for the single sign-on request: %s",
			response.StatusCode, string(body))
	}

	form, err := ParseAutoSubmitForm(string(body))
	if err != nil {
		return nil, fmt.Errorf("the identity provider returned no %s form: %s",
			SAMLResponseField, describeIdPResponse(response, string(body)))
	}

	samlResponse := form.Fields[SAMLResponseField]
	if samlResponse == "" {
		return nil, fmt.Errorf("the identity provider returned a form posting to %s rather than an assertion: %s",
			form.Action, describeIdPResponse(response, string(body)))
	}

	details, err := DecodeSAMLResponse(samlResponse)
	if err != nil {
		return nil, err
	}

	relayState := form.Fields[RelayStateField]
	if relayState == "" {
		relayState = loginRequest.RelayState
	}

	return &CapturedAssertion{
		Provider:     provider,
		SAMLResponse: samlResponse,
		RelayState:   relayState,
		StateCookies: loginRequest.StateCookies,
		Details:      details,
	}, nil
}

// SessionIdentity represents the user a Rancher session cookie authenticates as.
type SessionIdentity struct {
	UserID       string
	Username     string
	PrincipalIDs []string
}

// FetchSessionIdentity resolves a Rancher session cookie to the user it authenticates as
func FetchSessionIdentity(httpClient *http.Client, targetBaseURL, hostHeader, sessionToken string) (*SessionIdentity, error) {
	requestURL := strings.TrimRight(targetBaseURL, "/") + currentUserPath

	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building a session identity request to %s: %w", requestURL, err)
	}
	request.Header.Set("Accept", jsonContentType)
	if hostHeader != "" {
		request.Host = hostHeader
	}
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sessionToken})

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting the session identity from %s: %w", requestURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the session identity response from %s: %w", requestURL, err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the session was not accepted by %s, which returned %d: %s",
			requestURL, response.StatusCode, string(body))
	}

	payload := &struct {
		Data []struct {
			ID           string   `json:"id"`
			Username     string   `json:"username"`
			PrincipalIDs []string `json:"principalIds"`
		} `json:"data"`
	}{}
	if err := json.Unmarshal(body, payload); err != nil {
		return nil, fmt.Errorf("decoding the session identity response from %s, %s: %w", requestURL, string(body), err)
	}

	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("%s resolved the session to no user: %s", requestURL, string(body))
	}

	return &SessionIdentity{
		UserID:       payload.Data[0].ID,
		Username:     payload.Data[0].Username,
		PrincipalIDs: payload.Data[0].PrincipalIDs,
	}, nil
}

// SubmitAssertion presents a captured assertion to the assertion consumer service at targetBaseURL
func SubmitAssertion(httpClient *http.Client, targetBaseURL, hostHeader string, assertion *CapturedAssertion) (*ACSResult, error) {
	acsURL := strings.TrimRight(targetBaseURL, "/") + ACSPath(assertion.Provider)

	body := url.Values{
		SAMLResponseField: {assertion.SAMLResponse},
		RelayStateField:   {assertion.RelayState},
	}

	request, err := http.NewRequest(http.MethodPost, acsURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building an assertion consumer service request to %s: %w", acsURL, err)
	}
	request.Header.Set("Content-Type", formContentType)
	if hostHeader != "" {
		request.Host = hostHeader
	}
	for _, cookie := range assertion.StateCookies {
		request.AddCookie(cookie)
	}

	start := time.Now()
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("posting the assertion to %s: %w", acsURL, err)
	}
	defer response.Body.Close()
	latency := time.Since(start)

	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return nil, fmt.Errorf("reading the assertion consumer service response from %s: %w", acsURL, err)
	}

	result := &ACSResult{
		Target:     targetBaseURL,
		StatusCode: response.StatusCode,
		Location:   response.Header.Get("Location"),
		Latency:    latency,
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == SessionCookieName && cookie.Value != "" {
			result.SessionToken = cookie.Value
			result.Accepted = true
		}
	}

	return result, nil
}
