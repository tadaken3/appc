package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/tadaken3/appc/internal/client"
)

type AppResource struct {
	Type       string        `json:"type"`
	ID         string        `json:"id"`
	Attributes AppAttributes `json:"attributes"`
}

type AppAttributes struct {
	Name     string `json:"name"`
	BundleID string `json:"bundleId"`
	SKU      string `json:"sku"`
}

type App struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	BundleID    string  `json:"bundle_id"`
	SKU         string  `json:"sku"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"rating_count"`
}

func (a App) CSVHeaders() []string {
	return []string{"id", "name", "bundle_id", "sku", "rating", "rating_count"}
}

func (a App) CSVRow() []string {
	return []string{a.ID, a.Name, a.BundleID, a.SKU, fmt.Sprintf("%.2f", a.Rating), intToStr(a.RatingCount)}
}

func ListApps(ctx context.Context, c *client.Client) ([]App, error) {
	var apps []App
	url := "/v1/apps"

	for url != "" {
		resp, err := c.Get(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("listing apps: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[AppResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range result.Data {
			apps = append(apps, App{
				ID:       r.ID,
				Name:     r.Attributes.Name,
				BundleID: r.Attributes.BundleID,
				SKU:      r.Attributes.SKU,
			})
		}

		url = nextPath(result.Links.Next, c.BaseURL)
	}

	return apps, nil
}

// nextPath extracts the path from a full next URL, or returns "" if no next page.
func nextPath(nextURL, baseURL string) string {
	if nextURL == "" {
		return ""
	}
	if len(nextURL) > len(baseURL) && nextURL[:len(baseURL)] == baseURL {
		return nextURL[len(baseURL):]
	}
	return nextURL
}

// Helper used by multiple API packages
func intToStr(n int) string {
	return strconv.Itoa(n)
}
