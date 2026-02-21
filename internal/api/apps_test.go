package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenta-tanaka/appc/internal/client"
)

type stubTokenProvider struct{}

func (s *stubTokenProvider) Token() (string, error) { return "test-token", nil }

func TestListApps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apps" {
			t.Errorf("path = %q, want /v1/apps", r.URL.Path)
		}
		resp := Response[AppResource]{
			Data: []AppResource{
				{
					Type: "apps",
					ID:   "123",
					Attributes: AppAttributes{
						Name:     "My App",
						BundleID: "com.example.app",
						SKU:      "SKU001",
					},
				},
				{
					Type: "apps",
					ID:   "456",
					Attributes: AppAttributes{
						Name:     "Other App",
						BundleID: "com.example.other",
						SKU:      "SKU002",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	apps, err := ListApps(context.Background(), c)
	if err != nil {
		t.Fatalf("ListApps() error: %v", err)
	}

	if len(apps) != 2 {
		t.Fatalf("len = %d, want 2", len(apps))
	}
	if apps[0].ID != "123" {
		t.Errorf("apps[0].ID = %q, want 123", apps[0].ID)
	}
	if apps[0].Name != "My App" {
		t.Errorf("apps[0].Name = %q, want My App", apps[0].Name)
	}
	if apps[1].BundleID != "com.example.other" {
		t.Errorf("apps[1].BundleID = %q", apps[1].BundleID)
	}
}

func TestListAppsPagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[AppResource]
		if page == 1 {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "1", Type: "apps", Attributes: AppAttributes{Name: "App1"}},
				},
				Links: PagingLinks{Next: r.URL.Query().Get("") + "/v1/apps?cursor=page2"},
			}
		} else {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "2", Type: "apps", Attributes: AppAttributes{Name: "App2"}},
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Fix the next link to use the test server URL
	origHandler := srv.Config.Handler
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := 0
		origHandler.ServeHTTP(w, r)
		_ = page
	})

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	// Reset for actual test
	page = 0
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[AppResource]
		if page == 1 {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "1", Type: "apps", Attributes: AppAttributes{Name: "App1"}},
				},
				Links: PagingLinks{Next: srv.URL + "/v1/apps?cursor=page2"},
			}
		} else {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "2", Type: "apps", Attributes: AppAttributes{Name: "App2"}},
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	})

	apps, err := ListApps(context.Background(), c)
	if err != nil {
		t.Fatalf("ListApps() error: %v", err)
	}

	if len(apps) != 2 {
		t.Fatalf("len = %d, want 2", len(apps))
	}
	if apps[0].Name != "App1" || apps[1].Name != "App2" {
		t.Errorf("apps = %v", apps)
	}
}

func TestListAppsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response[AppResource]{Data: []AppResource{}})
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	apps, err := ListApps(context.Background(), c)
	if err != nil {
		t.Fatalf("ListApps() error: %v", err)
	}
	if len(apps) != 0 {
		t.Errorf("len = %d, want 0", len(apps))
	}
}
