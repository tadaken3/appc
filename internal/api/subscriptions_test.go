package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/kenta-tanaka/appc/internal/client"
)

func TestListSubscriptionGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/apps/APP123/subscriptionGroups"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}

		resp := Response[SubscriptionGroupResource]{
			Data: []SubscriptionGroupResource{
				{
					Type: "subscriptionGroups",
					ID:   "group1",
					Attributes: SubscriptionGroupAttributes{
						ReferenceName: "Premium Plans",
					},
				},
				{
					Type: "subscriptionGroups",
					ID:   "group2",
					Attributes: SubscriptionGroupAttributes{
						ReferenceName: "Basic Plans",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	groups, err := ListSubscriptionGroups(context.Background(), c, "APP123")
	if err != nil {
		t.Fatalf("ListSubscriptionGroups() error: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("len = %d, want 2", len(groups))
	}
	if groups[0].ID != "group1" {
		t.Errorf("groups[0].ID = %q, want group1", groups[0].ID)
	}
	if groups[0].ReferenceName != "Premium Plans" {
		t.Errorf("groups[0].ReferenceName = %q, want Premium Plans", groups[0].ReferenceName)
	}
	if groups[1].ID != "group2" {
		t.Errorf("groups[1].ID = %q, want group2", groups[1].ID)
	}
}

func TestListSubscriptionGroupsPagination(t *testing.T) {
	var page atomic.Int32
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := page.Add(1)
		var resp Response[SubscriptionGroupResource]
		if p == 1 {
			resp = Response[SubscriptionGroupResource]{
				Data: []SubscriptionGroupResource{
					{ID: "g1", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Group1"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/apps/APP1/subscriptionGroups?cursor=page2"},
			}
		} else {
			resp = Response[SubscriptionGroupResource]{
				Data: []SubscriptionGroupResource{
					{ID: "g2", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Group2"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	groups, err := ListSubscriptionGroups(context.Background(), c, "APP1")
	if err != nil {
		t.Fatalf("ListSubscriptionGroups() error: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("len = %d, want 2", len(groups))
	}
	if groups[0].ReferenceName != "Group1" || groups[1].ReferenceName != "Group2" {
		t.Errorf("groups = %v", groups)
	}
}

func TestListSubscriptionGroupsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response[SubscriptionGroupResource]{Data: []SubscriptionGroupResource{}})
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	groups, err := ListSubscriptionGroups(context.Background(), c, "APP1")
	if err != nil {
		t.Fatalf("ListSubscriptionGroups() error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("len = %d, want 0", len(groups))
	}
}

func TestListSubscriptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/subscriptionGroups/group1/subscriptions"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}

		resp := Response[SubscriptionResource]{
			Data: []SubscriptionResource{
				{
					Type: "subscriptions",
					ID:   "sub1",
					Attributes: SubscriptionAttributes{
						Name:                "Monthly Premium",
						ProductID:           "com.example.monthly",
						State:               "APPROVED",
						SubscriptionPeriod:  "ONE_MONTH",
						GroupLevel:          1,
						ReviewNote:          "Premium monthly plan",
						FamilySharable:      true,
						AvailableInAllTerritories: true,
					},
				},
				{
					Type: "subscriptions",
					ID:   "sub2",
					Attributes: SubscriptionAttributes{
						Name:                "Yearly Premium",
						ProductID:           "com.example.yearly",
						State:               "APPROVED",
						SubscriptionPeriod:  "ONE_YEAR",
						GroupLevel:          2,
						FamilySharable:      false,
						AvailableInAllTerritories: false,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	subs, err := ListSubscriptions(context.Background(), c, "group1")
	if err != nil {
		t.Fatalf("ListSubscriptions() error: %v", err)
	}

	if len(subs) != 2 {
		t.Fatalf("len = %d, want 2", len(subs))
	}
	if subs[0].SubscriptionID != "sub1" {
		t.Errorf("subs[0].SubscriptionID = %q, want sub1", subs[0].SubscriptionID)
	}
	if subs[0].Name != "Monthly Premium" {
		t.Errorf("subs[0].Name = %q, want Monthly Premium", subs[0].Name)
	}
	if subs[0].ProductID != "com.example.monthly" {
		t.Errorf("subs[0].ProductID = %q, want com.example.monthly", subs[0].ProductID)
	}
	if subs[0].State != "APPROVED" {
		t.Errorf("subs[0].State = %q, want APPROVED", subs[0].State)
	}
	if subs[0].SubscriptionPeriod != "ONE_MONTH" {
		t.Errorf("subs[0].SubscriptionPeriod = %q, want ONE_MONTH", subs[0].SubscriptionPeriod)
	}
	if subs[0].GroupLevel != 1 {
		t.Errorf("subs[0].GroupLevel = %d, want 1", subs[0].GroupLevel)
	}
	if subs[1].Name != "Yearly Premium" {
		t.Errorf("subs[1].Name = %q, want Yearly Premium", subs[1].Name)
	}
}

func TestListSubscriptionsPagination(t *testing.T) {
	var page atomic.Int32
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := page.Add(1)
		var resp Response[SubscriptionResource]
		if p == 1 {
			resp = Response[SubscriptionResource]{
				Data: []SubscriptionResource{
					{ID: "s1", Type: "subscriptions", Attributes: SubscriptionAttributes{Name: "Plan1", ProductID: "p1", State: "APPROVED"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/subscriptionGroups/g1/subscriptions?cursor=page2"},
			}
		} else {
			resp = Response[SubscriptionResource]{
				Data: []SubscriptionResource{
					{ID: "s2", Type: "subscriptions", Attributes: SubscriptionAttributes{Name: "Plan2", ProductID: "p2", State: "APPROVED"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	subs, err := ListSubscriptions(context.Background(), c, "g1")
	if err != nil {
		t.Fatalf("ListSubscriptions() error: %v", err)
	}

	if len(subs) != 2 {
		t.Fatalf("len = %d, want 2", len(subs))
	}
}

func TestListSubscriptionsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Response[SubscriptionResource]{Data: []SubscriptionResource{}})
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	subs, err := ListSubscriptions(context.Background(), c, "g1")
	if err != nil {
		t.Fatalf("ListSubscriptions() error: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("len = %d, want 0", len(subs))
	}
}

func TestFetchAppSubscriptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/apps/APP1/subscriptionGroups":
			resp := Response[SubscriptionGroupResource]{
				Data: []SubscriptionGroupResource{
					{ID: "g1", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Premium"}},
					{ID: "g2", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Basic"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/v1/subscriptionGroups/g1/subscriptions":
			resp := Response[SubscriptionResource]{
				Data: []SubscriptionResource{
					{ID: "s1", Type: "subscriptions", Attributes: SubscriptionAttributes{Name: "Monthly", ProductID: "com.example.monthly", State: "APPROVED"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/v1/subscriptionGroups/g2/subscriptions":
			resp := Response[SubscriptionResource]{
				Data: []SubscriptionResource{
					{ID: "s2", Type: "subscriptions", Attributes: SubscriptionAttributes{Name: "Free Trial", ProductID: "com.example.trial", State: "APPROVED"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	results, err := FetchAppSubscriptions(context.Background(), c, SubscriptionParams{AppID: "APP1"})
	if err != nil {
		t.Fatalf("FetchAppSubscriptions() error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	if results[0].GroupID != "g1" {
		t.Errorf("results[0].GroupID = %q, want g1", results[0].GroupID)
	}
	if results[0].GroupName != "Premium" {
		t.Errorf("results[0].GroupName = %q, want Premium", results[0].GroupName)
	}
	if results[0].SubscriptionID != "s1" {
		t.Errorf("results[0].SubscriptionID = %q, want s1", results[0].SubscriptionID)
	}
	if results[0].Name != "Monthly" {
		t.Errorf("results[0].Name = %q, want Monthly", results[0].Name)
	}
	if results[1].GroupName != "Basic" {
		t.Errorf("results[1].GroupName = %q, want Basic", results[1].GroupName)
	}
	if results[1].ProductID != "com.example.trial" {
		t.Errorf("results[1].ProductID = %q, want com.example.trial", results[1].ProductID)
	}
}

func TestFetchAppSubscriptionsWithGroupFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/apps/APP1/subscriptionGroups":
			resp := Response[SubscriptionGroupResource]{
				Data: []SubscriptionGroupResource{
					{ID: "g1", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Premium"}},
					{ID: "g2", Type: "subscriptionGroups", Attributes: SubscriptionGroupAttributes{ReferenceName: "Basic"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/v1/subscriptionGroups/g1/subscriptions":
			resp := Response[SubscriptionResource]{
				Data: []SubscriptionResource{
					{ID: "s1", Type: "subscriptions", Attributes: SubscriptionAttributes{Name: "Monthly", ProductID: "p1", State: "APPROVED"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			t.Errorf("unexpected path: %s (should not fetch g2)", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	results, err := FetchAppSubscriptions(context.Background(), c, SubscriptionParams{AppID: "APP1", GroupID: "g1"})
	if err != nil {
		t.Fatalf("FetchAppSubscriptions() error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].GroupID != "g1" {
		t.Errorf("results[0].GroupID = %q, want g1", results[0].GroupID)
	}
}

func TestSubscriptionInfoCSVOutput(t *testing.T) {
	s := SubscriptionInfo{
		GroupID:            "g1",
		GroupName:          "Premium",
		SubscriptionID:     "s1",
		Name:               "Monthly",
		ProductID:          "com.example.monthly",
		State:              "APPROVED",
		SubscriptionPeriod: "ONE_MONTH",
		GroupLevel:         1,
	}
	headers := s.CSVHeaders()
	row := s.CSVRow()
	if len(headers) != len(row) {
		t.Errorf("headers len = %d, row len = %d", len(headers), len(row))
	}
	expected := []string{"g1", "Premium", "s1", "Monthly", "com.example.monthly", "APPROVED", "ONE_MONTH", "1"}
	for i, want := range expected {
		if row[i] != want {
			t.Errorf("row[%d] (%s) = %q, want %q", i, headers[i], row[i], want)
		}
	}
}

func TestListSubscriptionGroupsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	_, err := ListSubscriptionGroups(context.Background(), c, "APP1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSubscriptionsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	_, err := ListSubscriptions(context.Background(), c, "g1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSubscriptionGroupsInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	_, err := ListSubscriptionGroups(context.Background(), c, "APP1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSubscriptionsInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	_, err := ListSubscriptions(context.Background(), c, "g1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchAppSubscriptionsEmptyAppID(t *testing.T) {
	c := client.New(&stubTokenProvider{})
	_, err := FetchAppSubscriptions(context.Background(), c, SubscriptionParams{AppID: ""})
	if err == nil {
		t.Fatal("expected error for empty AppID, got nil")
	}
}
