package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type mockTokenProvider struct {
	token string
}

func (m *mockTokenProvider) Token() (string, error) {
	return m.token, nil
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "test-token"})
	c.BaseURL = srv.URL

	_, err := c.Get(context.Background(), "/v1/apps")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-token")
	}
}

func TestRetryOn429(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL
	c.MaxRetries = 3

	resp, err := c.Get(context.Background(), "/v1/apps")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestRetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL
	c.MaxRetries = 2

	_, err := c.Get(context.Background(), "/v1/apps")
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
}

func TestErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors":[{"detail":"forbidden"}]}`))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL

	_, err := c.Get(context.Background(), "/v1/apps")
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %q, want it to contain 403", err.Error())
	}
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("raw data here"))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL

	resp, err := c.GetRaw(context.Background(), "/v1/salesReports?filter[reportType]=SALES")
	if err != nil {
		t.Fatalf("GetRaw() error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "raw data here" {
		t.Errorf("body = %q, want %q", body, "raw data here")
	}
}
