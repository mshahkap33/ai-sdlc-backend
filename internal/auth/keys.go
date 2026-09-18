package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ParseRSAPublicKeyFromPEM decodes a PEM-encoded RSA public key (either a
// PKIX "PUBLIC KEY" block or an X.509 certificate) as used to configure a
// Verifier from AuthConfig.JWTPublicKeyPEM.
func ParseRSAPublicKeyFromPEM(pemBytes string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		return nil, fmt.Errorf("auth: no PEM block found in public key")
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if key, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return key, nil
		}
		return nil, fmt.Errorf("auth: certificate public key is not RSA")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if key, ok := pub.(*rsa.PublicKey); ok {
			return key, nil
		}
		return nil, fmt.Errorf("auth: public key is not RSA")
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("auth: parsing RSA public key: %w", err)
	}
	return key, nil
}
