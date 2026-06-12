package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenLifetime = 20 * time.Minute

type TokenGenerator struct {
	issuerID string
	keyID    string
	key      *ecdsa.PrivateKey

	mu        sync.Mutex
	cached    string
	expiresAt time.Time
}

func NewTokenGenerator(issuerID, keyID, keyPath string) (*TokenGenerator, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}
	return NewTokenGeneratorFromPEM(issuerID, keyID, data)
}

// NewTokenGeneratorFromPEM builds a TokenGenerator from PEM-encoded private key
// bytes, allowing the key to be supplied from sources other than a file (e.g.
// an environment variable in a cloud environment).
func NewTokenGeneratorFromPEM(issuerID, keyID string, pemData []byte) (*TokenGenerator, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from private key")
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	ecKey, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not ECDSA")
	}

	return &TokenGenerator{
		issuerID: issuerID,
		keyID:    keyID,
		key:      ecKey,
	}, nil
}

func (g *TokenGenerator) Token() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.cached != "" && time.Now().Before(g.expiresAt) {
		return g.cached, nil
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    g.issuerID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
		Audience:  jwt.ClaimStrings{"appstoreconnect-v1"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = g.keyID

	signed, err := token.SignedString(g.key)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	g.cached = signed
	g.expiresAt = now.Add(tokenLifetime - 1*time.Minute) // refresh 1 min early
	return signed, nil
}
