package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/kenta-tanaka/appc/internal/client"
)

type ReviewResource struct {
	Type       string           `json:"type"`
	ID         string           `json:"id"`
	Attributes ReviewAttributes `json:"attributes"`
}

type ReviewAttributes struct {
	Rating      int    `json:"rating"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Reviewer    string `json:"reviewerNickname"`
	Territory   string `json:"territory"`
	CreatedDate string `json:"createdDate"`
}

type ReviewParams struct {
	AppID  string
	Rating int
	Limit  int
}

type ReviewSummary struct {
	Total    int            `json:"total"`
	Average  float64        `json:"average"`
	ByRating map[string]int `json:"by_rating"`
}

func (s ReviewSummary) CSVHeaders() []string {
	return []string{"total", "average", "rating_1", "rating_2", "rating_3", "rating_4", "rating_5"}
}

func (s ReviewSummary) CSVRow() []string {
	avg := fmt.Sprintf("%.2f", s.Average)
	return []string{
		intToStr(s.Total), avg,
		intToStr(s.ByRating["1"]), intToStr(s.ByRating["2"]), intToStr(s.ByRating["3"]),
		intToStr(s.ByRating["4"]), intToStr(s.ByRating["5"]),
	}
}

func SummarizeReviews(reviews []Review) ReviewSummary {
	byRating := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	total := len(reviews)
	sum := 0
	for _, r := range reviews {
		sum += r.Rating
		key := intToStr(r.Rating)
		byRating[key]++
	}
	avg := 0.0
	if total > 0 {
		avg = float64(sum) / float64(total)
	}
	return ReviewSummary{
		Total:    total,
		Average:  avg,
		ByRating: byRating,
	}
}

type Review struct {
	ID          string `json:"id"`
	Rating      int    `json:"rating"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	Reviewer    string `json:"reviewer"`
	Territory   string `json:"territory"`
	CreatedDate string `json:"created_date"`
}

func (r Review) CSVHeaders() []string {
	return []string{"id", "rating", "title", "body", "reviewer", "territory", "created_date"}
}

func (r Review) CSVRow() []string {
	return []string{r.ID, intToStr(r.Rating), r.Title, r.Body, r.Reviewer, r.Territory, r.CreatedDate}
}

func ListReviews(ctx context.Context, c *client.Client, params ReviewParams) ([]Review, error) {
	q := url.Values{}
	if params.Rating > 0 {
		q.Set("filter[rating]", strconv.Itoa(params.Rating))
	}
	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	}

	path := fmt.Sprintf("/v1/apps/%s/customerReviews", params.AppID)
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var reviews []Review
	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("listing reviews: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[ReviewResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range result.Data {
			reviews = append(reviews, Review{
				ID:          r.ID,
				Rating:      r.Attributes.Rating,
				Title:       r.Attributes.Title,
				Body:        r.Attributes.Body,
				Reviewer:    r.Attributes.Reviewer,
				Territory:   r.Attributes.Territory,
				CreatedDate: r.Attributes.CreatedDate,
			})
			if params.Limit > 0 && len(reviews) >= params.Limit {
				return reviews, nil
			}
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return reviews, nil
}
