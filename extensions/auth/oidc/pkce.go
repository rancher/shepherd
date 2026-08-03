package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type PKCEPair struct {
	Verifier  string
	Challenge string
}

// GeneratePKCE creates a fresh PKCE verifier/challenge pair using the S256 method.
func GeneratePKCE() (PKCEPair, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return PKCEPair{}, fmt.Errorf("generating PKCE verifier entropy: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return PKCEPair{Verifier: verifier, Challenge: challenge}, nil
}
