package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupAppRating(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("id") != "123456" {
			t.Errorf("id = %q, want 123456", q.Get("id"))
		}
		if q.Get("country") != "jp" {
			t.Errorf("country = %q, want jp", q.Get("country"))
		}
		resp := lookupResponse{
			ResultCount: 1,
			Results: []lookupResult{
				{
					TrackID:           123456,
					TrackName:         "My App",
					AverageUserRating: 4.81,
					UserRatingCount:   27,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	origBase := LookupBaseURL
	LookupBaseURL = srv.URL
	defer func() { LookupBaseURL = origBase }()

	rating, err := LookupAppRating(context.Background(), "123456", "jp")
	if err != nil {
		t.Fatalf("LookupAppRating() error: %v", err)
	}
	if rating.AppID != "123456" {
		t.Errorf("AppID = %q, want 123456", rating.AppID)
	}
	if rating.Name != "My App" {
		t.Errorf("Name = %q, want My App", rating.Name)
	}
	if rating.Rating != 4.81 {
		t.Errorf("Rating = %f, want 4.81", rating.Rating)
	}
	if rating.RatingCount != 27 {
		t.Errorf("RatingCount = %d, want 27", rating.RatingCount)
	}
}

func TestLookupAppRatingNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := lookupResponse{ResultCount: 0, Results: []lookupResult{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	origBase := LookupBaseURL
	LookupBaseURL = srv.URL
	defer func() { LookupBaseURL = origBase }()

	rating, err := LookupAppRating(context.Background(), "999999", "jp")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if rating.Rating != 0 || rating.RatingCount != 0 {
		t.Errorf("expected zero rating, got Rating=%f RatingCount=%d", rating.Rating, rating.RatingCount)
	}
}

func TestLookupAppRatingHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	origBase := LookupBaseURL
	LookupBaseURL = srv.URL
	defer func() { LookupBaseURL = origBase }()

	_, err := LookupAppRating(context.Background(), "123456", "jp")
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestLookupAppRatings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		switch id {
		case "111":
			resp := lookupResponse{
				ResultCount: 1,
				Results:     []lookupResult{{TrackID: 111, TrackName: "App One", AverageUserRating: 4.5, UserRatingCount: 100}},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "222":
			w.WriteHeader(http.StatusInternalServerError)
		case "333":
			resp := lookupResponse{
				ResultCount: 1,
				Results:     []lookupResult{{TrackID: 333, TrackName: "App Three", AverageUserRating: 3.0, UserRatingCount: 50}},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	origBase := LookupBaseURL
	LookupBaseURL = srv.URL
	defer func() { LookupBaseURL = origBase }()

	var warnings []string
	warn := func(id string, err error) {
		warnings = append(warnings, id)
	}

	ratings := LookupAppRatings(context.Background(), []string{"111", "222", "333"}, "jp", warn)

	if len(ratings) != 2 {
		t.Fatalf("len = %d, want 2", len(ratings))
	}
	if ratings["111"].Name != "App One" {
		t.Errorf("ratings[111].Name = %q", ratings["111"].Name)
	}
	if ratings["333"].Name != "App Three" {
		t.Errorf("ratings[333].Name = %q", ratings["333"].Name)
	}
	if len(warnings) != 1 || warnings[0] != "222" {
		t.Errorf("warnings = %v, want [222]", warnings)
	}
}
