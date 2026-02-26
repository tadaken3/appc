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
	"strings"
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

// Instance types
type AnalyticsInstanceResource struct {
	Type       string                        `json:"type"`
	ID         string                        `json:"id"`
	Attributes AnalyticsInstanceAttributes   `json:"attributes"`
}

type AnalyticsInstanceAttributes struct {
	Granularity    string `json:"granularity"`
	ProcessingDate string `json:"processingDate"`
}

// Output type for instances
type AnalyticsInstance struct {
	ID             string `json:"id"`
	Granularity    string `json:"granularity"`
	ProcessingDate string `json:"processing_date"`
}

func (i AnalyticsInstance) CSVHeaders() []string {
	return []string{"id", "granularity", "processing_date"}
}

func (i AnalyticsInstance) CSVRow() []string {
	return []string{i.ID, i.Granularity, i.ProcessingDate}
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

// AnalyticsSegmentRecord represents a row from a downloaded analytics segment CSV.
// Headers are dynamic (vary by report category), so we use a map.
type AnalyticsSegmentRecord struct {
	Headers []string          `json:"-"`
	Fields  map[string]string `json:"fields"`
}

func (r AnalyticsSegmentRecord) CSVHeaders() []string {
	return r.Headers
}

func (r AnalyticsSegmentRecord) CSVRow() []string {
	row := make([]string, len(r.Headers))
	for i, h := range r.Headers {
		row[i] = r.Fields[h]
	}
	return row
}

func DownloadAnalyticsSegmentCSV(ctx context.Context, segmentURL string) ([]AnalyticsSegmentRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, segmentURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading segment CSV: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("segment download error %d", resp.StatusCode)
	}

	var reader io.Reader

	// Apple's segment URLs often return gzip-compressed data.
	// Check Content-Encoding header first, then try to detect gzip magic bytes.
	contentEncoding := resp.Header.Get("Content-Encoding")
	contentType := resp.Header.Get("Content-Type")
	if contentEncoding == "gzip" || contentType == "application/gzip" || contentType == "application/x-gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing segment CSV: %w", err)
		}
		defer func() { _ = gz.Close() }()
		reader = gz
	} else {
		// Try gzip decompression by reading initial bytes
		buf := make([]byte, 2)
		n, err := io.ReadFull(resp.Body, buf)
		if err != nil && n == 0 {
			return nil, nil
		}
		combined := io.MultiReader(bytes.NewReader(buf[:n]), resp.Body)
		// gzip magic number: 0x1f 0x8b
		if n == 2 && buf[0] == 0x1f && buf[1] == 0x8b {
			gz, err := gzip.NewReader(combined)
			if err != nil {
				return nil, fmt.Errorf("decompressing segment CSV: %w", err)
			}
			defer func() { _ = gz.Close() }()
			reader = gz
		} else {
			reader = combined
		}
	}

	return parseAnalyticsCSV(reader)
}

func parseAnalyticsCSV(r io.Reader) ([]AnalyticsSegmentRecord, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	lines, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	if len(lines) < 2 {
		return nil, nil
	}

	headers := lines[0]
	var records []AnalyticsSegmentRecord
	for _, fields := range lines[1:] {
		if len(fields) == 0 || (len(fields) == 1 && strings.TrimSpace(fields[0]) == "") {
			continue
		}
		m := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(fields) {
				m[h] = fields[i]
			}
		}
		records = append(records, AnalyticsSegmentRecord{
			Headers: headers,
			Fields:  m,
		})
	}

	return records, nil
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

	// 409 Conflict means an ONGOING request already exists; retrieve it
	if resp.StatusCode == http.StatusConflict {
		return getExistingReportRequestID(ctx, c, appID)
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

func getExistingReportRequestID(ctx context.Context, c *client.Client, appID string) (string, error) {
	path := fmt.Sprintf("/v1/apps/%s/analyticsReportRequests?filter[accessType]=ONGOING", appID)

	resp, err := c.Get(ctx, path)
	if err != nil {
		return "", fmt.Errorf("getting existing report requests: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var result Response[AnalyticsReportRequestResource]
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("no existing ONGOING analytics report request found")
	}

	return result.Data[0].ID, nil
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

func GetAnalyticsInstances(ctx context.Context, c *client.Client, reportID string) ([]AnalyticsInstance, error) {
	path := fmt.Sprintf("/v1/analyticsReports/%s/instances", reportID)

	var instances []AnalyticsInstance
	for path != "" {
		resp, err := c.Get(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("getting analytics instances: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		var result Response[AnalyticsInstanceResource]
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		for _, inst := range result.Data {
			instances = append(instances, AnalyticsInstance{
				ID:             inst.ID,
				Granularity:    inst.Attributes.Granularity,
				ProcessingDate: inst.Attributes.ProcessingDate,
			})
		}

		path = nextPath(result.Links.Next, c.BaseURL)
	}

	return instances, nil
}

func GetAnalyticsSegments(ctx context.Context, c *client.Client, instanceID string) ([]AnalyticsSegment, error) {
	path := fmt.Sprintf("/v1/analyticsReportInstances/%s/segments", instanceID)

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
