package client

import (
	"bytes"
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
		_, _ = w.Write([]byte(`{}`))
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
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL
	c.MaxRetries = 3

	resp, err := c.Get(context.Background(), "/v1/apps")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

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

func TestErrorResponseIncludesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"detail":"forbidden"}]}`))
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
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("error = %q, want it to contain response body", err.Error())
	}
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("raw data here"))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL

	resp, err := c.GetRaw(context.Background(), "/v1/salesReports?filter[reportType]=SALES")
	if err != nil {
		t.Fatalf("GetRaw() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "raw data here" {
		t.Errorf("body = %q, want %q", body, "raw data here")
	}
}

func TestPostRetryPreservesBody(t *testing.T) {
	var attempts int32
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastBody = string(body)
		n := atomic.AddInt32(&attempts, 1)
		if n < 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"123"}}`))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL
	c.MaxRetries = 3

	reqBody := []byte(`{"data":{"type":"test"}}`)
	resp, err := c.Post(context.Background(), "/v1/test", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
	if lastBody != `{"data":{"type":"test"}}` {
		t.Errorf("retried body = %q, want original body preserved", lastBody)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestGetRawWithAccept(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gzip data"))
	}))
	defer srv.Close()

	c := New(&mockTokenProvider{token: "tok"})
	c.BaseURL = srv.URL

	resp, err := c.GetRawWithAccept(context.Background(), "/v1/salesReports", "application/a-gzip")
	if err != nil {
		t.Fatalf("GetRawWithAccept() error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if gotAccept != "application/a-gzip" {
		t.Errorf("Accept = %q, want %q", gotAccept, "application/a-gzip")
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "gzip data" {
		t.Errorf("body = %q, want %q", body, "gzip data")
	}
}

func TestRetryAfterParsesSeconds(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int // seconds
	}{
		{"zero", "0", 0},
		{"five", "5", 5},
		{"empty", "", 1},
		{"invalid", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryAfterDuration(tt.value)
			wantDuration := tt.want * int(1e9) // convert to nanoseconds
			if int(got) != wantDuration {
				t.Errorf("retryAfterDuration(%q) = %v, want %ds", tt.value, got, tt.want)
			}
		})
	}
}
