package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestP8Key(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling key: %v", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	path := filepath.Join(t.TempDir(), "test.p8")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating file: %v", err)
	}
	defer f.Close()
	if err := pem.Encode(f, block); err != nil {
		t.Fatalf("encoding pem: %v", err)
	}
	return path
}

func TestGenerateToken(t *testing.T) {
	keyPath := generateTestP8Key(t)

	gen, err := NewTokenGenerator("issuer-123", "key-456", keyPath)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error: %v", err)
	}

	token, err := gen.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}

	if token == "" {
		t.Fatal("Token() returned empty string")
	}

	// Parse and verify token claims
	parsed, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		return &gen.key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("parsing token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("token is not valid")
	}

	if parsed.Method.Alg() != "ES256" {
		t.Errorf("algorithm = %q, want ES256", parsed.Method.Alg())
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims")
	}

	if iss, _ := claims.GetIssuer(); iss != "issuer-123" {
		t.Errorf("iss = %q, want %q", iss, "issuer-123")
	}

	if aud, _ := claims.GetAudience(); len(aud) == 0 || aud[0] != "appstoreconnect-v1" {
		t.Errorf("aud = %v, want [appstoreconnect-v1]", aud)
	}
}

func TestTokenCaching(t *testing.T) {
	keyPath := generateTestP8Key(t)

	gen, err := NewTokenGenerator("issuer-123", "key-456", keyPath)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error: %v", err)
	}

	token1, err := gen.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}

	token2, err := gen.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}

	if token1 != token2 {
		t.Error("expected cached token to be returned")
	}
}

func TestTokenExpiry(t *testing.T) {
	keyPath := generateTestP8Key(t)

	gen, err := NewTokenGenerator("issuer-123", "key-456", keyPath)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error: %v", err)
	}

	token1, _ := gen.Token()

	// Force expiry
	gen.expiresAt = time.Now().Add(-1 * time.Minute)

	token2, _ := gen.Token()

	if token1 == token2 {
		t.Error("expected new token after expiry")
	}
}

func TestInvalidKeyPath(t *testing.T) {
	_, err := NewTokenGenerator("issuer", "key", "/nonexistent/key.p8")
	if err == nil {
		t.Fatal("expected error for invalid key path")
	}
}
