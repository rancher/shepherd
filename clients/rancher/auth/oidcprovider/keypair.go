package oidcprovider

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
	keyPairBits       = 2048
	keyPairLifetime   = 365 * 24 * time.Hour
	certificatePEM    = "CERTIFICATE"
	privateKeyPEMType = "RSA PRIVATE KEY"
)

type keyPair struct {
	Certificate string
	PrivateKey  string
}

func newClientKeyPair(commonName string) (*keyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyPairBits)
	if err != nil {
		return nil, fmt.Errorf("generating a client key for %s: %w", commonName, err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, fmt.Errorf("generating a serial number for the %s client certificate: %w", commonName, err)
	}

	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(keyPairLifetime),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("creating the %s client certificate: %w", commonName, err)
	}

	return &keyPair{
		Certificate: string(pem.EncodeToMemory(&pem.Block{Type: certificatePEM, Bytes: der})),
		PrivateKey:  string(pem.EncodeToMemory(&pem.Block{Type: privateKeyPEMType, Bytes: x509.MarshalPKCS1PrivateKey(key)})),
	}, nil
}
