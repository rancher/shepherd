package saml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"time"
)

const samlTimeLayout = time.RFC3339

var (
	formBlockPattern  = regexp.MustCompile(`(?is)<form\b.*?</form>`)
	formActionPattern = regexp.MustCompile(`(?is)<form[^>]*\saction=["']([^"']+)["']`)
	formInputPattern  = regexp.MustCompile(`(?is)<input[^>]*>`)
	inputNamePattern  = regexp.MustCompile(`(?is)\sname=["']([^"']*)["']`)
	inputValuePattern = regexp.MustCompile(`(?is)\svalue=["']([^"']*)["']`)
)

// AutoSubmitForm represents the self-posting form an identity provider returns to deliver a SAML response.
type AutoSubmitForm struct {
	Action string
	Fields map[string]string
}

// AssertionDetails represents the identity, attributes and validity bounds carried by a SAML assertion.
type AssertionDetails struct {
	ID           string
	NameID       string
	Attributes   map[string][]string
	IssueInstant time.Time
	NotBefore    time.Time
	NotOnOrAfter time.Time
}

// Attribute returns the values an assertion carries for an attribute
func (d *AssertionDetails) Attribute(name string) []string {
	return d.Attributes[name]
}

type assertionDocument struct {
	XMLName   xml.Name `xml:"Response"`
	Assertion struct {
		ID           string `xml:"ID,attr"`
		IssueInstant string `xml:"IssueInstant,attr"`
		Conditions   struct {
			NotBefore    string `xml:"NotBefore,attr"`
			NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
		} `xml:"Conditions"`
		Subject struct {
			NameID string `xml:"NameID"`
		} `xml:"Subject"`
		AttributeStatement struct {
			Attributes []struct {
				Name         string   `xml:"Name,attr"`
				FriendlyName string   `xml:"FriendlyName,attr"`
				Values       []string `xml:"AttributeValue"`
			} `xml:"Attribute"`
		} `xml:"AttributeStatement"`
	} `xml:"Assertion"`
	EncryptedAssertion struct {
		XMLName xml.Name
	} `xml:"EncryptedAssertion"`
}

// ParseAutoSubmitForm extracts the action and fields of the form that carries a SAML response
func ParseAutoSubmitForm(document string) (*AutoSubmitForm, error) {
	forms := []*AutoSubmitForm{}
	for _, block := range formBlockPattern.FindAllString(document, -1) {
		form := parseForm(block)
		if form == nil {
			continue
		}

		if form.Fields[SAMLResponseField] != "" {
			return form, nil
		}

		forms = append(forms, form)
	}

	if len(forms) > 0 {
		return forms[0], nil
	}

	if form := parseForm(document); form != nil {
		return form, nil
	}

	return nil, fmt.Errorf("the document contains no HTML form with an action")
}

func parseForm(block string) *AutoSubmitForm {
	actionMatch := formActionPattern.FindStringSubmatch(block)
	if actionMatch == nil {
		return nil
	}

	form := &AutoSubmitForm{
		Action: html.UnescapeString(actionMatch[1]),
		Fields: map[string]string{},
	}

	for _, inputTag := range formInputPattern.FindAllString(block, -1) {
		nameMatch := inputNamePattern.FindStringSubmatch(inputTag)
		valueMatch := inputValuePattern.FindStringSubmatch(inputTag)
		if nameMatch == nil || valueMatch == nil {
			continue
		}

		form.Fields[html.UnescapeString(nameMatch[1])] = html.UnescapeString(valueMatch[1])
	}

	return form
}

// DecodeSAMLResponse decodes a SAML response and returns the assertion details it carries
func DecodeSAMLResponse(encodedResponse string) (*AssertionDetails, error) {
	rawResponse, err := base64.StdEncoding.DecodeString(encodedResponse)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding the SAML response: %w", err)
	}

	document := &assertionDocument{}
	if err := xml.Unmarshal(rawResponse, document); err != nil {
		return nil, fmt.Errorf("XML decoding the SAML response: %w", err)
	}

	if document.Assertion.ID == "" {
		if document.EncryptedAssertion.XMLName.Local != "" {
			return nil, fmt.Errorf("the SAML response carries an encrypted assertion, which this client cannot read")
		}

		return nil, fmt.Errorf("the SAML response carries no assertion ID")
	}

	details := &AssertionDetails{
		ID:         document.Assertion.ID,
		NameID:     document.Assertion.Subject.NameID,
		Attributes: map[string][]string{},
	}

	for _, attribute := range document.Assertion.AttributeStatement.Attributes {
		for _, key := range []string{attribute.Name, attribute.FriendlyName} {
			if key == "" {
				continue
			}

			details.Attributes[key] = append(details.Attributes[key], attribute.Values...)
		}
	}

	timestamps := []struct {
		name  string
		value string
		field *time.Time
	}{
		{"IssueInstant", document.Assertion.IssueInstant, &details.IssueInstant},
		{"Conditions NotBefore", document.Assertion.Conditions.NotBefore, &details.NotBefore},
		{"Conditions NotOnOrAfter", document.Assertion.Conditions.NotOnOrAfter, &details.NotOnOrAfter},
	}
	for _, timestamp := range timestamps {
		if timestamp.value == "" {
			continue
		}

		parsed, err := time.Parse(samlTimeLayout, timestamp.value)
		if err != nil {
			return nil, fmt.Errorf("parsing the assertion %s %q: %w", timestamp.name, timestamp.value, err)
		}

		*timestamp.field = parsed
	}

	return details, nil
}
