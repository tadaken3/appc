package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kenta-tanaka/appc/internal/client"
)

// Report Request types
type AnalyticsReportRequestResource struct {
	Type       string                             `json:"type"`
	ID         string                             `json:"id"`
	Attributes AnalyticsReportRequestAttributes   `json:"attributes"`
}

type AnalyticsReportRequestAttributes struct {
	AccessType string `json:"accessType"`
}

// Report types
type AnalyticsReportResource struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes AnalyticsReportAttributes   `json:"attributes"`
}

type AnalyticsReportAttributes struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// Segment types
type AnalyticsSegmentResource struct {
	Type       string                       `json:"type"`
	ID         string                       `json:"id"`
	Attributes AnalyticsSegmentAttributes   `json:"attributes"`
}

type AnalyticsSegmentAttributes struct {
	URL         string `json:"url"`
	CheckSum    string `json:"checkSum"`
	SizeInBytes int64  `json:"sizeInBytes"`
}

// Output types
type AnalyticsReport struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Name     string `json:"name"`
}

func (r AnalyticsReport) CSVHeaders() []string {
	return []string{"id", "category", "name"}
}

func (r AnalyticsReport) CSVRow() []string {
	return []string{r.ID, r.Category, r.Name}
}

type AnalyticsSegment struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	CheckSum    string `json:"checksum"`
	SizeInBytes int64  `json:"size_in_bytes"`
}

func (s AnalyticsSegment) CSVHeaders() []string {
	return []string{"id", "url", "checksum", "size_in_bytes"}
}

func (s AnalyticsSegment) CSVRow() []string {
	return []string{s.ID, s.URL, s.CheckSum, fmt.Sprintf("%d", s.SizeInBytes)}
}

func RequestAnalyticsReport(ctx context.Context, c *client.Client, appID string) (string, error) {
	reqBody := map[string]any{
		"data": map[string]any{
			"type": "analyticsReportRequests",
			"attributes": map[string]any{
				"accessType": "ONGOING",
			},
			"relationships": map[string]any{
				"app": map[string]any{
					"data": map[string]any{
						"type": "apps",
						"id":   appID,
					},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.Post(ctx, "/v1/analyticsReportRequests", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("requesting analytics report: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("analytics request API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result SingleResponse[AnalyticsReportRequestResource]
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return result.Data.ID, nil
}

func GetAnalyticsReports(ctx context.Context, c *client.Client, requestID, category string) ([]AnalyticsReport, error) {
	q := url.Values{}
	if category != "" {
		q.Set("filter[category]", category)
	}

	path := fmt.Sprintf("/v1/analyticsReportRequests/%s/reports", requestID)
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("getting analytics reports: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result Response[AnalyticsReportResource]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var reports []AnalyticsReport
	for _, r := range result.Data {
		reports = append(reports, AnalyticsReport{
			ID:       r.ID,
			Category: r.Attributes.Category,
			Name:     r.Attributes.Name,
		})
	}

	return reports, nil
}

func GetAnalyticsSegments(ctx context.Context, c *client.Client, reportID string) ([]AnalyticsSegment, error) {
	path := fmt.Sprintf("/v1/analyticsReports/%s/segments", reportID)

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("getting analytics segments: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result Response[AnalyticsSegmentResource]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var segments []AnalyticsSegment
	for _, s := range result.Data {
		segments = append(segments, AnalyticsSegment{
			ID:          s.ID,
			URL:         s.Attributes.URL,
			CheckSum:    s.Attributes.CheckSum,
			SizeInBytes: s.Attributes.SizeInBytes,
		})
	}

	return segments, nil
}
