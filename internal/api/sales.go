package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/kenta-tanaka/appc/internal/client"
)

type SalesReportParams struct {
	ReportType   string
	ReportDate   string
	Frequency    string
	ReportSubType string
	VendorNumber string
}

type SalesRecord struct {
	Provider          string `json:"provider"`
	ProviderCountry   string `json:"provider_country"`
	SKU               string `json:"sku"`
	Developer         string `json:"developer"`
	Title             string `json:"title"`
	Version           string `json:"version"`
	ProductType       string `json:"product_type"`
	Units             string `json:"units"`
	DeveloperProceeds string `json:"developer_proceeds"`
	BeginDate         string `json:"begin_date"`
	EndDate           string `json:"end_date"`
	CustomerCurrency  string `json:"customer_currency"`
	CountryCode       string `json:"country_code"`
	CurrencyProceeds  string `json:"currency_of_proceeds"`
	AppleIdentifier   string `json:"apple_identifier"`
	CustomerPrice     string `json:"customer_price"`
	PromoCode         string `json:"promo_code"`
	ParentIdentifier  string `json:"parent_identifier"`
	Subscription      string `json:"subscription"`
	Period            string `json:"period"`
	Category          string `json:"category"`
	CMB               string `json:"cmb"`
	Device            string `json:"device"`
	SupportedPlatforms string `json:"supported_platforms"`
	ProceedsReason    string `json:"proceeds_reason"`
	PreservedPricing  string `json:"preserved_pricing"`
	Client            string `json:"client"`
	OrderType         string `json:"order_type"`
}

func (r SalesRecord) CSVHeaders() []string {
	return []string{
		"provider", "provider_country", "sku", "developer", "title", "version",
		"product_type", "units", "developer_proceeds", "begin_date", "end_date",
		"customer_currency", "country_code", "currency_of_proceeds", "apple_identifier",
		"customer_price", "promo_code", "parent_identifier", "subscription", "period",
		"category", "cmb", "device", "supported_platforms", "proceeds_reason",
		"preserved_pricing", "client", "order_type",
	}
}

func (r SalesRecord) CSVRow() []string {
	return []string{
		r.Provider, r.ProviderCountry, r.SKU, r.Developer, r.Title, r.Version,
		r.ProductType, r.Units, r.DeveloperProceeds, r.BeginDate, r.EndDate,
		r.CustomerCurrency, r.CountryCode, r.CurrencyProceeds, r.AppleIdentifier,
		r.CustomerPrice, r.PromoCode, r.ParentIdentifier, r.Subscription, r.Period,
		r.Category, r.CMB, r.Device, r.SupportedPlatforms, r.ProceedsReason,
		r.PreservedPricing, r.Client, r.OrderType,
	}
}

func GetSalesReport(ctx context.Context, c *client.Client, params SalesReportParams) ([]SalesRecord, error) {
	q := url.Values{}
	q.Set("filter[reportType]", params.ReportType)
	q.Set("filter[reportDate]", params.ReportDate)
	q.Set("filter[frequency]", params.Frequency)
	q.Set("filter[vendorNumber]", params.VendorNumber)
	if params.ReportSubType != "" {
		q.Set("filter[reportSubType]", params.ReportSubType)
	} else {
		q.Set("filter[reportSubType]", "SUMMARY")
	}

	path := "/v1/salesReports?" + q.Encode()

	resp, err := c.GetRawWithAccept(ctx, path, "application/a-gzip")
	if err != nil {
		return nil, fmt.Errorf("fetching sales report: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sales report API error %d: %s", resp.StatusCode, string(body))
	}

	var reader io.Reader
	contentEncoding := resp.Header.Get("Content-Encoding")
	contentType := resp.Header.Get("Content-Type")
	if contentEncoding == "gzip" || contentType == "application/gzip" || contentType == "application/x-gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing response: %w", err)
		}
		defer func() { _ = gz.Close() }()
		reader = gz
	} else {
		// Detect gzip by magic bytes (0x1f 0x8b)
		buf := make([]byte, 2)
		n, err := io.ReadFull(resp.Body, buf)
		if n == 0 {
			if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, fmt.Errorf("reading response preamble: %w", err)
			}
			return nil, nil
		}
		combined := io.MultiReader(bytes.NewReader(buf[:n]), resp.Body)
		if n == 2 && buf[0] == 0x1f && buf[1] == 0x8b {
			gz, err := gzip.NewReader(combined)
			if err != nil {
				return nil, fmt.Errorf("decompressing response: %w", err)
			}
			defer func() { _ = gz.Close() }()
			reader = gz
		} else {
			reader = combined
		}
	}

	return parseTSV(reader)
}

func parseTSV(r io.Reader) ([]SalesRecord, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1

	lines, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading TSV: %w", err)
	}

	if len(lines) < 2 {
		return nil, nil
	}

	// Build header-to-index map
	headerIdx := make(map[string]int, len(lines[0]))
	for i, h := range lines[0] {
		headerIdx[h] = i
	}

	field := func(fields []string, header string) string {
		if idx, ok := headerIdx[header]; ok && idx < len(fields) {
			return fields[idx]
		}
		return ""
	}

	var records []SalesRecord
	for _, fields := range lines[1:] {
		if len(fields) == 0 || (len(fields) == 1 && strings.TrimSpace(fields[0]) == "") {
			continue
		}
		records = append(records, SalesRecord{
			Provider:           field(fields, "Provider"),
			ProviderCountry:    field(fields, "Provider Country"),
			SKU:                field(fields, "SKU"),
			Developer:          field(fields, "Developer"),
			Title:              field(fields, "Title"),
			Version:            field(fields, "Version"),
			ProductType:        field(fields, "Product Type Identifier"),
			Units:              field(fields, "Units"),
			DeveloperProceeds:  field(fields, "Developer Proceeds"),
			BeginDate:          field(fields, "Begin Date"),
			EndDate:            field(fields, "End Date"),
			CustomerCurrency:   field(fields, "Customer Currency"),
			CountryCode:        field(fields, "Country Code"),
			CurrencyProceeds:   field(fields, "Currency of Proceeds"),
			AppleIdentifier:    field(fields, "Apple Identifier"),
			CustomerPrice:      field(fields, "Customer Price"),
			PromoCode:          field(fields, "Promo Code"),
			ParentIdentifier:   field(fields, "Parent Identifier"),
			Subscription:       field(fields, "Subscription"),
			Period:             field(fields, "Period"),
			Category:           field(fields, "Category"),
			CMB:                field(fields, "CMB"),
			Device:             field(fields, "Device"),
			SupportedPlatforms: field(fields, "Supported Platforms"),
			ProceedsReason:     field(fields, "Proceeds Reason"),
			PreservedPricing:   field(fields, "Preserved Pricing"),
			Client:             field(fields, "Client"),
			OrderType:          field(fields, "Order Type"),
		})
	}

	return records, nil
}
