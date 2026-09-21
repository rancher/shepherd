package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

const (
	keyPairBits     = 2048
	keyPairLifetime = 365 * 24 * time.Hour
)

// KeyPair represents a service provider signing certificate and its private key, both PEM encoded.
type KeyPair struct {
	Certificate string
	PrivateKey  string
}

// NewKeyPair generates a self signed service provider signing key pair
func NewKeyPair(commonName string) (*KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyPairBits)
	if err != nil {
		return nil, fmt.Errorf("generating a service provider signing key: %w", err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, fmt.Errorf("generating a certificate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(keyPairLifetime),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("generating a service provider signing certificate: %w", err)
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return &KeyPair{
		Certificate: string(certificatePEM),
		PrivateKey:  string(keyPEM),
	}, nil
}
