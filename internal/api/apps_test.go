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
		_ = json.NewEncoder(w).Encode(resp)
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
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[AppResource]
		if page == 1 {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "1", Type: "apps", Attributes: AppAttributes{Name: "App1"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/apps?cursor=page2"},
			}
		} else {
			resp = Response[AppResource]{
				Data: []AppResource{
					{ID: "2", Type: "apps", Attributes: AppAttributes{Name: "App2"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

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
		_ = json.NewEncoder(w).Encode(Response[AppResource]{Data: []AppResource{}})
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

func TestAppCSVOutput(t *testing.T) {
	a := App{
		ID:          "123",
		Name:        "My App",
		BundleID:    "com.example.app",
		SKU:         "SKU001",
		Rating:      4.81,
		RatingCount: 27,
	}
	headers := a.CSVHeaders()
	row := a.CSVRow()
	if len(headers) != 6 {
		t.Errorf("headers len = %d, want 6", len(headers))
	}
	if len(row) != len(headers) {
		t.Errorf("row len = %d, headers len = %d", len(row), len(headers))
	}
	if row[4] != "4.81" {
		t.Errorf("row[4] (rating) = %q, want 4.81", row[4])
	}
	if row[5] != "27" {
		t.Errorf("row[5] (rating_count) = %q, want 27", row[5])
	}
}
