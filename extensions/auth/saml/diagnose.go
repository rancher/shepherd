package saml

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const IdPResponseDumpDirEnvVar = "SAML_IDP_DUMP_DIR"

var pageTitlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func describeIdPResponse(response *http.Response, body string) string {
	finalURL := ""
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}

	description := []string{
		fmt.Sprintf("the identity provider answered %d and left the client at %s", response.StatusCode, finalURL),
	}

	if title := pageTitle(body); title != "" {
		description = append(description, fmt.Sprintf("the page is titled %q", title))
	}

	if dumpPath := dumpIdPResponse(body); dumpPath != "" {
		description = append(description, "the full response was written to "+dumpPath)
	} else {
		description = append(description, "set "+IdPResponseDumpDirEnvVar+" to a directory to capture the full response")
	}

	return strings.Join(description, "; ")
}

func pageTitle(body string) string {
	titleMatch := pageTitlePattern.FindStringSubmatch(body)
	if titleMatch == nil {
		return ""
	}

	return strings.TrimSpace(titleMatch[1])
}

func dumpIdPResponse(body string) string {
	dumpDir := os.Getenv(IdPResponseDumpDirEnvVar)
	if dumpDir == "" {
		return ""
	}

	if err := os.MkdirAll(dumpDir, 0o755); err != nil {
		return ""
	}

	dumpPath := filepath.Join(dumpDir, fmt.Sprintf("idp-response-%d.html", time.Now().UnixNano()))
	if err := os.WriteFile(dumpPath, []byte(body), 0o600); err != nil {
		return ""
	}

	return dumpPath
}
