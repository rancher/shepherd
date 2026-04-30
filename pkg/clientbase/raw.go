package clientbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// Do executes a raw HTTP request; body may be url.Values (form-encoded) or any JSON-serializable value.
func Do(client *http.Client, method, rawURL string, body interface{}, headers map[string]string) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case url.Values:
			bodyReader = strings.NewReader(v.Encode())
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("marshaling request body: %w", err)
			}
			bodyReader = bytes.NewReader(b)
		}
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request %s %s: %w", method, rawURL, err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request %s %s: %w", method, rawURL, err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body) //nolint:errcheck
		resp.Body.Close()              //nolint:errcheck
	}()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body from %s %s: %w", method, rawURL, err)
	}
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Header:     resp.Header,
	}, nil
}
