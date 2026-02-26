package api

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kenta-tanaka/appc/internal/client"
)

func TestRequestAnalyticsReport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/analyticsReportRequests" {
			t.Errorf("path = %q", r.URL.Path)
		}

		var reqBody struct {
			Data struct {
				Type       string `json:"type"`
				Attributes struct {
					AccessType string `json:"accessType"`
				} `json:"attributes"`
				Relationships struct {
					App struct {
						Data struct {
							Type string `json:"type"`
							ID   string `json:"id"`
						} `json:"data"`
					} `json:"app"`
				} `json:"relationships"`
			} `json:"data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)

		if reqBody.Data.Relationships.App.Data.ID != "APP123" {
			t.Errorf("appID = %q", reqBody.Data.Relationships.App.Data.ID)
		}

		resp := SingleResponse[AnalyticsReportRequestResource]{
			Data: AnalyticsReportRequestResource{
				Type: "analyticsReportRequests",
				ID:   "req-001",
				Attributes: AnalyticsReportRequestAttributes{
					AccessType: "ONGOING",
				},
			},
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reqID, err := RequestAnalyticsReport(context.Background(), c, "APP123")
	if err != nil {
		t.Fatalf("RequestAnalyticsReport() error: %v", err)
	}
	if reqID != "req-001" {
		t.Errorf("reqID = %q, want req-001", reqID)
	}
}

func TestGetAnalyticsReports(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/v1/analyticsReportRequests/req-001/reports"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}

		resp := Response[AnalyticsReportResource]{
			Data: []AnalyticsReportResource{
				{
					Type: "analyticsReports",
					ID:   "report-1",
					Attributes: AnalyticsReportAttributes{
						Category: "APP_USAGE",
						Name:     "App Usage Report",
					},
				},
				{
					Type: "analyticsReports",
					ID:   "report-2",
					Attributes: AnalyticsReportAttributes{
						Category: "APP_STORE_ENGAGEMENT",
						Name:     "Engagement Report",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reports, err := GetAnalyticsReports(context.Background(), c, "req-001", "")
	if err != nil {
		t.Fatalf("GetAnalyticsReports() error: %v", err)
	}

	if len(reports) != 2 {
		t.Fatalf("len = %d, want 2", len(reports))
	}
	if reports[0].Category != "APP_USAGE" {
		t.Errorf("category = %q", reports[0].Category)
	}
}

func TestGetAnalyticsReportsWithCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cat := r.URL.Query().Get("filter[category]"); cat != "APP_USAGE" {
			t.Errorf("filter[category] = %q, want APP_USAGE", cat)
		}
		resp := Response[AnalyticsReportResource]{
			Data: []AnalyticsReportResource{
				{ID: "r1", Type: "analyticsReports", Attributes: AnalyticsReportAttributes{Category: "APP_USAGE"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reports, err := GetAnalyticsReports(context.Background(), c, "req-001", "APP_USAGE")
	if err != nil {
		t.Fatalf("GetAnalyticsReports() error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("len = %d, want 1", len(reports))
	}
}

func TestGetAnalyticsSegments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response[AnalyticsSegmentResource]{
			Data: []AnalyticsSegmentResource{
				{
					Type: "analyticsReportSegments",
					ID:   "seg-1",
					Attributes: AnalyticsSegmentAttributes{
						URL:       "https://example.com/data.csv",
						CheckSum:  "abc123",
						SizeInBytes: 1024,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	segments, err := GetAnalyticsSegments(context.Background(), c, "report-1")
	if err != nil {
		t.Fatalf("GetAnalyticsSegments() error: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("len = %d, want 1", len(segments))
	}
	if segments[0].URL != "https://example.com/data.csv" {
		t.Errorf("URL = %q", segments[0].URL)
	}
}

func TestAnalyticsReportCSVOutput(t *testing.T) {
	r := AnalyticsReport{
		ID:       "r1",
		Category: "APP_USAGE",
		Name:     "Usage",
	}
	headers := r.CSVHeaders()
	row := r.CSVRow()
	if len(headers) != len(row) {
		t.Errorf("headers len = %d, row len = %d", len(headers), len(row))
	}
}

func TestFetchAnalyticsFlow(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/analyticsReportRequests":
			resp := SingleResponse[AnalyticsReportRequestResource]{
				Data: AnalyticsReportRequestResource{ID: "req-1", Type: "analyticsReportRequests"},
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/v1/analyticsReportRequests/req-1/reports":
			resp := Response[AnalyticsReportResource]{
				Data: []AnalyticsReportResource{
					{ID: "rep-1", Type: "analyticsReports", Attributes: AnalyticsReportAttributes{Category: "APP_USAGE", Name: "Usage"}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			t.Logf("call %d: %s %s", n, r.Method, r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"data":[]}`)
		}
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reqID, err := RequestAnalyticsReport(context.Background(), c, "APP1")
	if err != nil {
		t.Fatalf("request error: %v", err)
	}

	reports, err := GetAnalyticsReports(context.Background(), c, reqID, "APP_USAGE")
	if err != nil {
		t.Fatalf("reports error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("len = %d, want 1", len(reports))
	}
}

func TestGetAnalyticsReportsPagination(t *testing.T) {
	page := 0
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[AnalyticsReportResource]
		if page == 1 {
			resp = Response[AnalyticsReportResource]{
				Data: []AnalyticsReportResource{
					{ID: "r1", Type: "analyticsReports", Attributes: AnalyticsReportAttributes{Category: "APP_USAGE", Name: "Page1"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/analyticsReportRequests/req-1/reports?cursor=2"},
			}
		} else {
			resp = Response[AnalyticsReportResource]{
				Data: []AnalyticsReportResource{
					{ID: "r2", Type: "analyticsReports", Attributes: AnalyticsReportAttributes{Category: "APP_USAGE", Name: "Page2"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	reports, err := GetAnalyticsReports(context.Background(), c, "req-1", "")
	if err != nil {
		t.Fatalf("GetAnalyticsReports() error: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("len = %d, want 2", len(reports))
	}
	if reports[0].Name != "Page1" || reports[1].Name != "Page2" {
		t.Errorf("reports = %v", reports)
	}
}

func TestDownloadAnalyticsSegmentCSV(t *testing.T) {
	csvData := "Date,App Name,Downloads\n2026-02-20,MyApp,42\n2026-02-21,MyApp,55\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = fmt.Fprint(w, csvData)
	}))
	defer srv.Close()

	records, err := DownloadAnalyticsSegmentCSV(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("DownloadAnalyticsSegmentCSV() error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len = %d, want 2", len(records))
	}
	if records[0].Fields["Date"] != "2026-02-20" {
		t.Errorf("Date = %q, want 2026-02-20", records[0].Fields["Date"])
	}
	if records[0].Fields["Downloads"] != "42" {
		t.Errorf("Downloads = %q, want 42", records[0].Fields["Downloads"])
	}
	if records[1].Fields["App Name"] != "MyApp" {
		t.Errorf("App Name = %q, want MyApp", records[1].Fields["App Name"])
	}
}

func TestDownloadAnalyticsSegmentCSVGzip(t *testing.T) {
	csvData := "Date,Downloads\n2026-02-20,42\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = fmt.Fprint(gz, csvData)
		_ = gz.Close()
	}))
	defer srv.Close()

	records, err := DownloadAnalyticsSegmentCSV(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("DownloadAnalyticsSegmentCSV() error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}
	if records[0].Fields["Date"] != "2026-02-20" {
		t.Errorf("Date = %q", records[0].Fields["Date"])
	}
}

func TestDownloadAnalyticsSegmentCSVEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Date,Downloads\n")
	}))
	defer srv.Close()

	records, err := DownloadAnalyticsSegmentCSV(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("DownloadAnalyticsSegmentCSV() error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("len = %d, want 0", len(records))
	}
}

func TestDownloadAnalyticsSegmentCSVHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := DownloadAnalyticsSegmentCSV(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want to contain 404", err.Error())
	}
}

func TestAnalyticsSegmentRecordCSVOutput(t *testing.T) {
	headers := []string{"Date", "App Name", "Downloads"}
	r := AnalyticsSegmentRecord{
		Headers: headers,
		Fields:  map[string]string{"Date": "2026-02-20", "App Name": "MyApp", "Downloads": "42"},
	}
	csvHeaders := r.CSVHeaders()
	if len(csvHeaders) != 3 {
		t.Fatalf("CSVHeaders len = %d, want 3", len(csvHeaders))
	}
	row := r.CSVRow()
	if len(row) != 3 {
		t.Fatalf("CSVRow len = %d, want 3", len(row))
	}
	if row[0] != "2026-02-20" {
		t.Errorf("row[0] = %q, want 2026-02-20", row[0])
	}
}

func TestGetAnalyticsSegmentsPagination(t *testing.T) {
	page := 0
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		var resp Response[AnalyticsSegmentResource]
		if page == 1 {
			resp = Response[AnalyticsSegmentResource]{
				Data: []AnalyticsSegmentResource{
					{ID: "s1", Type: "analyticsReportSegments", Attributes: AnalyticsSegmentAttributes{URL: "https://example.com/1.csv"}},
				},
				Links: PagingLinks{Next: srvURL + "/v1/analyticsReports/rep-1/segments?cursor=2"},
			}
		} else {
			resp = Response[AnalyticsSegmentResource]{
				Data: []AnalyticsSegmentResource{
					{ID: "s2", Type: "analyticsReportSegments", Attributes: AnalyticsSegmentAttributes{URL: "https://example.com/2.csv"}},
				},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	srvURL = srv.URL

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	segments, err := GetAnalyticsSegments(context.Background(), c, "rep-1")
	if err != nil {
		t.Fatalf("GetAnalyticsSegments() error: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len = %d, want 2", len(segments))
	}
}
