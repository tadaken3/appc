package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenta-tanaka/appc/internal/client"
)

func TestListReviews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/apps/APP123/customerReviews"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}

		resp := Response[ReviewResource]{
			Data: []ReviewResource{
				{
					Type: "customerReviews",
					ID:   "rev1",
					Attributes: ReviewAttributes{
						Rating:    5,
						Title:     "Great app!",
						Body:      "Love it",
						Reviewer:  "user1",
						Territory: "US",
						CreatedDate: "2025-01-15T10:00:00Z",
					},
				},
				{
					Type: "customerReviews",
					ID:   "rev2",
					Attributes: ReviewAttributes{
						Rating:    3,
						Title:     "OK",
						Body:      "Could be better",
						Reviewer:  "user2",
						Territory: "JP",
						CreatedDate: "2025-01-14T08:00:00Z",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	params := ReviewParams{
		AppID: "APP123",
		Limit: 100,
	}

	reviews, err := ListReviews(context.Background(), c, params)
	if err != nil {
		t.Fatalf("ListReviews() error: %v", err)
	}

	if len(reviews) != 2 {
		t.Fatalf("len = %d, want 2", len(reviews))
	}

	if reviews[0].Rating != 5 {
		t.Errorf("reviews[0].Rating = %d, want 5", reviews[0].Rating)
	}
	if reviews[0].Title != "Great app!" {
		t.Errorf("reviews[0].Title = %q", reviews[0].Title)
	}
	if reviews[1].Territory != "JP" {
		t.Errorf("reviews[1].Territory = %q", reviews[1].Territory)
	}
}

func TestListReviewsWithRatingFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if rating := q.Get("filter[rating]"); rating != "5" {
			t.Errorf("filter[rating] = %q, want 5", rating)
		}
		resp := Response[ReviewResource]{
			Data: []ReviewResource{
				{
					Type: "customerReviews",
					ID:   "rev1",
					Attributes: ReviewAttributes{Rating: 5, Title: "Perfect"},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	params := ReviewParams{AppID: "APP1", Rating: 5}
	reviews, err := ListReviews(context.Background(), c, params)
	if err != nil {
		t.Fatalf("ListReviews() error: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("len = %d, want 1", len(reviews))
	}
}

func TestListReviewsPagination(t *testing.T) {
	page := 0
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[ReviewResource]
		if page == 1 {
			resp = Response[ReviewResource]{
				Data: []ReviewResource{
					{ID: "r1", Type: "customerReviews", Attributes: ReviewAttributes{Title: "First"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/apps/APP1/customerReviews?cursor=2"},
			}
		} else {
			resp = Response[ReviewResource]{
				Data: []ReviewResource{
					{ID: "r2", Type: "customerReviews", Attributes: ReviewAttributes{Title: "Second"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reviews, err := ListReviews(context.Background(), c, ReviewParams{AppID: "APP1"})
	if err != nil {
		t.Fatalf("ListReviews() error: %v", err)
	}

	if len(reviews) != 2 {
		t.Fatalf("len = %d, want 2", len(reviews))
	}
}

func TestListReviewsLimitStopsPagination(t *testing.T) {
	page := 0
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := Response[ReviewResource]{
			Data: []ReviewResource{
				{ID: "r1", Type: "customerReviews", Attributes: ReviewAttributes{Title: "Review"}},
				{ID: "r2", Type: "customerReviews", Attributes: ReviewAttributes{Title: "Review2"}},
			},
			Links: PagingLinks{Next: srvURL + "/v1/apps/APP1/customerReviews?cursor=next"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reviews, err := ListReviews(context.Background(), c, ReviewParams{AppID: "APP1", Limit: 1})
	if err != nil {
		t.Fatalf("ListReviews() error: %v", err)
	}

	if len(reviews) != 1 {
		t.Errorf("len = %d, want 1 (should stop at limit)", len(reviews))
	}
	if page != 1 {
		t.Errorf("pages fetched = %d, want 1 (should not fetch next page)", page)
	}
}

func TestReviewCSVOutput(t *testing.T) {
	r := Review{
		ID:       "r1",
		Rating:   5,
		Title:    "Good",
		Body:     "Nice",
		Reviewer: "user",
	}
	headers := r.CSVHeaders()
	row := r.CSVRow()
	if len(headers) != len(row) {
		t.Errorf("headers len = %d, row len = %d", len(headers), len(row))
	}
}

func TestSummarizeReviews(t *testing.T) {
	reviews := []Review{
		{Rating: 5},
		{Rating: 5},
		{Rating: 4},
		{Rating: 3},
		{Rating: 1},
	}

	s := SummarizeReviews(reviews)

	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}

	// average = (5+5+4+3+1)/5 = 3.6
	if s.Average < 3.59 || s.Average > 3.61 {
		t.Errorf("Average = %.2f, want 3.60", s.Average)
	}

	if s.ByRating["5"] != 2 {
		t.Errorf("ByRating[5] = %d, want 2", s.ByRating["5"])
	}
	if s.ByRating["4"] != 1 {
		t.Errorf("ByRating[4] = %d, want 1", s.ByRating["4"])
	}
	if s.ByRating["3"] != 1 {
		t.Errorf("ByRating[3] = %d, want 1", s.ByRating["3"])
	}
	if s.ByRating["2"] != 0 {
		t.Errorf("ByRating[2] = %d, want 0", s.ByRating["2"])
	}
	if s.ByRating["1"] != 1 {
		t.Errorf("ByRating[1] = %d, want 1", s.ByRating["1"])
	}
}

func TestSummarizeReviewsEmpty(t *testing.T) {
	s := SummarizeReviews([]Review{})
	if s.Total != 0 {
		t.Errorf("Total = %d, want 0", s.Total)
	}
	if s.Average != 0.0 {
		t.Errorf("Average = %.2f, want 0.00", s.Average)
	}
}

func TestReviewSummaryCSVOutput(t *testing.T) {
	s := ReviewSummary{
		Total:    5,
		Average:  3.6,
		ByRating: map[string]int{"1": 1, "2": 0, "3": 1, "4": 1, "5": 2},
	}
	headers := s.CSVHeaders()
	row := s.CSVRow()
	if len(headers) != len(row) {
		t.Errorf("headers len = %d, row len = %d", len(headers), len(row))
	}
	if row[0] != "5" {
		t.Errorf("row[0] (total) = %q, want 5", row[0])
	}
	if row[1] != "3.60" {
		t.Errorf("row[1] (average) = %q, want 3.60", row[1])
	}
}
