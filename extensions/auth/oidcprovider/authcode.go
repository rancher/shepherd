package oidcprovider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	samlext "github.com/rancher/shepherd/extensions/auth/saml"
	"github.com/rancher/shepherd/extensions/defaults"
)

const (
	codeParam             = "code"
	stateParam            = "state"
	errorParam            = "error"
	errorDescriptionParam = "error_description"
	responseTypeCode      = "code"

	formContentType = "application/x-www-form-urlencoded"
	maxRedirects    = 15
)

// AuthorizationURL builds the authorization code request an identity provider is asked to answer
func AuthorizationURL(authorizationEndpoint, clientID, redirectURI, scopes, state string) (string, error) {
	endpoint, err := url.Parse(authorizationEndpoint)
	if err != nil {
		return "", fmt.Errorf("parsing the authorization endpoint %s: %w", authorizationEndpoint, err)
	}

	query := endpoint.Query()
	query.Set("response_type", responseTypeCode)
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", scopes)
	query.Set(stateParam, state)

	endpoint.RawQuery = query.Encode()

	return endpoint.String(), nil
}

// CaptureAuthorizationCode signs a user in at the identity provider and returns the code it redirects back with
func CaptureAuthorizationCode(transport http.RoundTripper, authorizationURL, redirectURI, username, password string) (string, error) {
	requested, err := url.Parse(authorizationURL)
	if err != nil {
		return "", fmt.Errorf("parsing the authorization request %s: %w", authorizationURL, err)
	}

	expectedState := requested.Query().Get(stateParam)

	callback, err := url.Parse(redirectURI)
	if err != nil {
		return "", fmt.Errorf("parsing the redirect URI %s: %w", redirectURI, err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("creating a cookie jar for the identity provider session: %w", err)
	}

	var redirected *url.URL

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   defaults.OneMinuteTimeout,
		Jar:       jar,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if isCallback(request.URL, callback) {
				redirected = request.URL

				return http.ErrUseLastResponse
			}

			if len(via) >= maxRedirects {
				return fmt.Errorf("the identity provider redirected %d times without reaching %s", len(via), redirectURI)
			}

			return nil
		},
	}

	response, document, err := getDocument(httpClient, authorizationURL)
	if err != nil {
		return "", err
	}

	if redirected != nil {
		return codeFrom(redirected, expectedState, username)
	}

	loginForm, err := samlext.ParseLoginForm(document)
	if err != nil {
		return "", fmt.Errorf("the identity provider served no sign-in form to %s at %s, it answered %d: %s",
			username, authorizationURL, response.StatusCode, truncate(document))
	}

	action, err := resolveReference(response.Request.URL, loginForm.Action)
	if err != nil {
		return "", fmt.Errorf("resolving the sign-in form action %q for %s: %w", loginForm.Action, username, err)
	}

	credentials := url.Values{}
	for name, value := range loginForm.Fields {
		credentials.Set(name, value)
	}
	if loginForm.UsernameField != "" {
		credentials.Set(loginForm.UsernameField, username)
	}
	credentials.Set(loginForm.PasswordField, password)

	postResponse, postDocument, err := postForm(httpClient, action, credentials, response.Request.URL.String())
	if err != nil {
		return "", fmt.Errorf("submitting the sign-in form for %s: %w", username, err)
	}

	if redirected != nil {
		return codeFrom(redirected, expectedState, username)
	}

	if _, stillAsking := samlext.ParseLoginForm(postDocument); stillAsking == nil {
		return "", fmt.Errorf("the identity provider returned its sign-in form again for %s, which means the "+
			"credentials were rejected", username)
	}

	return "", fmt.Errorf("the identity provider never redirected %s back to %s after the sign-in form was "+
		"submitted, it answered %d: %s", username, redirectURI, postResponse.StatusCode, truncate(postDocument))
}

// isCallback reports whether a redirect landed on the configured callback address
func isCallback(candidate, callback *url.URL) bool {
	return strings.EqualFold(candidate.Scheme, callback.Scheme) &&
		strings.EqualFold(candidate.Host, callback.Host) &&
		candidate.Path == callback.Path
}

// codeFrom reads the authorization code out of the callback the identity provider redirected to
func codeFrom(redirected *url.URL, expectedState, username string) (string, error) {
	query := redirected.Query()

	if failure := query.Get(errorParam); failure != "" {
		return "", fmt.Errorf("the identity provider refused to issue a code for %s, it answered %s: %s",
			username, failure, query.Get(errorDescriptionParam))
	}

	if returnedState := query.Get(stateParam); returnedState != expectedState {
		return "", fmt.Errorf("the identity provider answered the authorization request for %s with state %q "+
			"rather than the %q it was given, so the code it carries answers some other request",
			username, returnedState, expectedState)
	}

	code := query.Get(codeParam)
	if code == "" {
		return "", fmt.Errorf("the identity provider redirected %s to %s carrying no authorization code",
			username, redirected.Path)
	}

	return code, nil
}

// resolveReference turns a form action into an absolute address against the page it was served on
func resolveReference(pageURL *url.URL, action string) (string, error) {
	if action == "" {
		return pageURL.String(), nil
	}

	reference, err := url.Parse(action)
	if err != nil {
		return "", err
	}

	return pageURL.ResolveReference(reference).String(), nil
}

// getDocument fetches a page from the identity provider and returns the response alongside its body
func getDocument(httpClient *http.Client, requestURL string) (*http.Response, string, error) {
	response, err := httpClient.Get(requestURL)
	if err != nil {
		return nil, "", fmt.Errorf("requesting %s: %w", requestURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading the response from %s: %w", requestURL, err)
	}

	return response, string(body), nil
}

// postForm submits a form to the identity provider and returns the response alongside its body
func postForm(httpClient *http.Client, requestURL string, values url.Values, referer string) (*http.Response, string, error) {
	request, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("building a form post to %s: %w", requestURL, err)
	}

	request.Header.Set("Content-Type", formContentType)
	if referer != "" {
		request.Header.Set("Referer", referer)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("posting the form to %s: %w", requestURL, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading the response from %s: %w", requestURL, err)
	}

	return response, string(body), nil
}

// truncate shortens a page body so it can be named in an error
func truncate(document string) string {
	const limit = 512

	if len(document) <= limit {
		return document
	}

	return document[:limit] + "..."
}
