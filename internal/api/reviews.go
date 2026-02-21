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
		resp.Body.Close()
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
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return reviews, nil
}
