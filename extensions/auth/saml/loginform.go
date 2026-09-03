package saml

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	passwordInputType = "password"
	hiddenInputType   = "hidden"
	submitInputType   = "submit"
	textInputType     = "text"
	emailInputType    = "email"
)

var inputTypePattern = regexp.MustCompile(`(?is)\stype=["']([^"']*)["']`)

// LoginForm represents an identity provider's HTML sign-in form.
type LoginForm struct {
	Action        string
	UsernameField string
	PasswordField string
	Fields        map[string]string
}

// AuthenticateWithLoginForm signs a user in at an identity provider that authenticates with an HTML form
func AuthenticateWithLoginForm(httpClient *http.Client, startURL, username, password string) (string, error) {
	response, body, err := getDocument(httpClient, startURL)
	if err != nil {
		return "", err
	}

	if carriesAssertion(body) {
		return body, nil
	}

	loginForm, err := ParseLoginForm(body)
	if err != nil {
		return "", fmt.Errorf("signing %s in at the identity provider: %s",
			username, describeIdPResponse(response, body))
	}

	actionURL, err := resolveReference(response.Request.URL, loginForm.Action)
	if err != nil {
		return "", fmt.Errorf("resolving the sign-in form action %q for %s: %w", loginForm.Action, username, err)
	}

	credentials := url.Values{}
	for name, value := range loginForm.Fields {
		credentials.Set(name, value)
	}
	credentials.Set(loginForm.UsernameField, username)
	credentials.Set(loginForm.PasswordField, password)

	postResponse, postBody, err := postForm(httpClient, actionURL, credentials, response.Request.URL.String())
	if err != nil {
		return "", fmt.Errorf("submitting the sign-in form for %s: %w", username, err)
	}

	if carriesAssertion(postBody) {
		return postBody, nil
	}

	if _, stillAsking := ParseLoginForm(postBody); stillAsking == nil {
		return "", fmt.Errorf("the identity provider returned its sign-in form again for %s, which means the "+
			"credentials were rejected", username)
	}

	return "", fmt.Errorf("the identity provider did not issue an assertion for %s after the sign-in form was "+
		"submitted, which usually means it asked for an additional factor: %s",
		username, describeIdPResponse(postResponse, postBody))
}

// ParseLoginForm finds the sign-in form in an identity provider page
func ParseLoginForm(document string) (*LoginForm, error) {
	for _, block := range formBlockPattern.FindAllString(document, -1) {
		actionMatch := formActionPattern.FindStringSubmatch(block)
		if actionMatch == nil {
			continue
		}

		loginForm := &LoginForm{
			Action: html.UnescapeString(actionMatch[1]),
			Fields: map[string]string{},
		}

		for _, inputTag := range formInputPattern.FindAllString(block, -1) {
			nameMatch := inputNamePattern.FindStringSubmatch(inputTag)
			if nameMatch == nil {
				continue
			}
			name := html.UnescapeString(nameMatch[1])
			if name == "" {
				continue
			}

			inputType := textInputType
			if typeMatch := inputTypePattern.FindStringSubmatch(inputTag); typeMatch != nil {
				inputType = strings.ToLower(html.UnescapeString(typeMatch[1]))
			}

			value := ""
			if valueMatch := inputValuePattern.FindStringSubmatch(inputTag); valueMatch != nil {
				value = html.UnescapeString(valueMatch[1])
			}

			switch inputType {
			case passwordInputType:
				if loginForm.PasswordField == "" {
					loginForm.PasswordField = name
				}
			case textInputType, emailInputType:
				if loginForm.UsernameField == "" {
					loginForm.UsernameField = name
				}
			case hiddenInputType, submitInputType:
				loginForm.Fields[name] = value
			}
		}

		if loginForm.PasswordField == "" || loginForm.UsernameField == "" {
			continue
		}

		return loginForm, nil
	}

	return nil, fmt.Errorf("the document contains no sign-in form with a username and a password input")
}

func carriesAssertion(document string) bool {
	form, err := ParseAutoSubmitForm(document)

	return err == nil && form.Fields[SAMLResponseField] != ""
}

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
