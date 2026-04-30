package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// DecodeJWTPayload decodes the payload section of a JWT without verifying the signature.
func DecodeJWTPayload(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT structure: expected 3 parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decoding JWT payload: %w", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parsing JWT claims: %w", err)
	}
	return claims, nil
}

// TamperJWTSignature replaces the signature segment of a JWT with random bytes.
func TamperJWTSignature(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT: cannot tamper signature")
	}
	fakeSig := make([]byte, 32)
	if _, err := rand.Read(fakeSig); err != nil {
		return "", fmt.Errorf("generating fake signature: %w", err)
	}
	return parts[0] + "." + parts[1] + "." + base64.RawURLEncoding.EncodeToString(fakeSig), nil
}
