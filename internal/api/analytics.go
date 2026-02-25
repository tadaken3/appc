package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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

// AnalyticsDataRecord represents a row from an analytics segment CSV.
// Headers are dynamic and depend on the report category.
type AnalyticsDataRecord struct {
	headers []string
	fields  []string
}

func (r AnalyticsDataRecord) CSVHeaders() []string {
	return r.headers
}

func (r AnalyticsDataRecord) CSVRow() []string {
	return r.fields
}

func (r AnalyticsDataRecord) MarshalJSON() ([]byte, error) {
	m := make(map[string]string, len(r.headers))
	for i, h := range r.headers {
		if i < len(r.fields) {
			m[h] = r.fields[i]
		}
	}
	return json.Marshal(m)
}

// ParseAnalyticsCSV parses a CSV reader into AnalyticsDataRecord slices.
func ParseAnalyticsCSV(r io.Reader) ([]AnalyticsDataRecord, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true

	lines, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	if len(lines) < 2 {
		return nil, nil
	}

	headers := lines[0]
	var records []AnalyticsDataRecord
	for _, fields := range lines[1:] {
		records = append(records, AnalyticsDataRecord{
			headers: headers,
			fields:  fields,
		})
	}

	return records, nil
}

// DownloadSegmentCSV downloads a segment CSV from the given URL and parses it.
func DownloadSegmentCSV(ctx context.Context, segmentURL string) ([]AnalyticsDataRecord, error) {
	httpClient := &http.Client{Timeout: 60 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, segmentURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading segment CSV: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("segment download error: status %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing segment CSV: %w", err)
		}
		defer func() { _ = gz.Close() }()
		reader = gz
	}

	return ParseAnalyticsCSV(reader)
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
	defer func() { _ = resp.Body.Close() }()

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

	var reports []AnalyticsReport
	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("getting analytics reports: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[AnalyticsReportResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, r := range result.Data {
			reports = append(reports, AnalyticsReport{
				ID:       r.ID,
				Category: r.Attributes.Category,
				Name:     r.Attributes.Name,
			})
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return reports, nil
}

func GetAnalyticsSegments(ctx context.Context, c *client.Client, reportID string) ([]AnalyticsSegment, error) {
	path := fmt.Sprintf("/v1/analyticsReports/%s/segments", reportID)

	var segments []AnalyticsSegment
	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("getting analytics segments: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[AnalyticsSegmentResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, s := range result.Data {
			segments = append(segments, AnalyticsSegment{
				ID:          s.ID,
				URL:         s.Attributes.URL,
				CheckSum:    s.Attributes.CheckSum,
				SizeInBytes: s.Attributes.SizeInBytes,
			})
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return segments, nil
}
