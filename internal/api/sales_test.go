package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kenta-tanaka/appc/internal/client"
)

func gzipTSV(t *testing.T, data string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	gz.Close()
	return buf.Bytes()
}

func TestGetSalesReport(t *testing.T) {
	tsvData := "Provider\tProvider Country\tSKU\tDeveloper\tTitle\tVersion\tProduct Type Identifier\tUnits\tDeveloper Proceeds\tBegin Date\tEnd Date\tCustomer Currency\tCountry Code\tCurrency of Proceeds\tApple Identifier\tCustomer Price\tPromo Code\tParent Identifier\tSubscription\tPeriod\tCategory\tCMB\tDevice\tSupported Platforms\tProceeds Reason\tPreserved Pricing\tClient\tOrder Type\n" +
		"APPLE\tUS\tSKU001\tDev\tMy App\t1.0\t1F\t10\t7.00\t01/15/2025\t01/15/2025\tUSD\tUS\tUSD\t123456\t0.99\t\t\t\t\tGames\t\tiPhone\t\t\t\t\t\n" +
		"APPLE\tJP\tSKU001\tDev\tMy App\t1.0\tIA1\t5\t490\t01/15/2025\t01/15/2025\tJPY\tJP\tJPY\t123456\t800\t\t\t\t\tGames\t\tiPhone\t\t\t\t\t\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/salesReports" {
			t.Errorf("path = %q, want /v1/salesReports", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("filter[reportType]") != "SALES" {
			t.Errorf("reportType = %q", q.Get("filter[reportType]"))
		}
		if q.Get("filter[reportDate]") != "2025-01-15" {
			t.Errorf("reportDate = %q", q.Get("filter[reportDate]"))
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(gzipTSV(t, tsvData))
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	params := SalesReportParams{
		ReportType: "SALES",
		ReportDate: "2025-01-15",
		Frequency:  "DAILY",
		VendorNumber: "12345",
	}

	records, err := GetSalesReport(context.Background(), c, params)
	if err != nil {
		t.Fatalf("GetSalesReport() error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("len = %d, want 2", len(records))
	}

	r := records[0]
	if r.SKU != "SKU001" {
		t.Errorf("SKU = %q", r.SKU)
	}
	if r.Title != "My App" {
		t.Errorf("Title = %q", r.Title)
	}
	if r.Units != "10" {
		t.Errorf("Units = %q, want 10", r.Units)
	}
	if r.CountryCode != "US" {
		t.Errorf("CountryCode = %q", r.CountryCode)
	}

	r2 := records[1]
	if r2.CountryCode != "JP" {
		t.Errorf("records[1].CountryCode = %q", r2.CountryCode)
	}
	if r2.DeveloperProceeds != "490" {
		t.Errorf("DeveloperProceeds = %q", r2.DeveloperProceeds)
	}
}

func TestGetSalesReportCSVOutput(t *testing.T) {
	r := SalesRecord{
		Provider:    "APPLE",
		SKU:         "SKU001",
		Title:       "Test",
		Units:       "10",
		CountryCode: "US",
	}
	headers := r.CSVHeaders()
	if len(headers) == 0 {
		t.Fatal("CSVHeaders() returned empty")
	}
	row := r.CSVRow()
	if len(row) != len(headers) {
		t.Errorf("CSVRow len = %d, headers len = %d", len(row), len(headers))
	}
}

func TestSalesReportEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		// Just the header, no data rows
		tsvData := "Provider\tProvider Country\tSKU\tDeveloper\tTitle\tVersion\tProduct Type Identifier\tUnits\tDeveloper Proceeds\tBegin Date\tEnd Date\tCustomer Currency\tCountry Code\tCurrency of Proceeds\tApple Identifier\tCustomer Price\tPromo Code\tParent Identifier\tSubscription\tPeriod\tCategory\tCMB\tDevice\tSupported Platforms\tProceeds Reason\tPreserved Pricing\tClient\tOrder Type\n"
		w.Write(gzipTSV(t, tsvData))
	}))
	defer srv.Close()

	c := client.New(&stubTokenProvider{})
	c.BaseURL = srv.URL

	records, err := GetSalesReport(context.Background(), c, SalesReportParams{
		ReportType: "SALES", ReportDate: "2025-01-15", Frequency: "DAILY", VendorNumber: "12345",
	})
	if err != nil {
		t.Fatalf("GetSalesReport() error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("len = %d, want 0", len(records))
	}
}
