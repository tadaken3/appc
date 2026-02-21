package api

import (
	"compress/gzip"
	"context"
	"encoding/csv"
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

	resp, err := c.GetRaw(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetching sales report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sales report API error %d: %s", resp.StatusCode, string(body))
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing response: %w", err)
		}
		defer gz.Close()
		reader = gz
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

	var records []SalesRecord
	for _, fields := range lines[1:] {
		// Skip empty lines
		if len(fields) == 0 || (len(fields) == 1 && strings.TrimSpace(fields[0]) == "") {
			continue
		}
		rec := SalesRecord{}
		if len(fields) > 0 { rec.Provider = fields[0] }
		if len(fields) > 1 { rec.ProviderCountry = fields[1] }
		if len(fields) > 2 { rec.SKU = fields[2] }
		if len(fields) > 3 { rec.Developer = fields[3] }
		if len(fields) > 4 { rec.Title = fields[4] }
		if len(fields) > 5 { rec.Version = fields[5] }
		if len(fields) > 6 { rec.ProductType = fields[6] }
		if len(fields) > 7 { rec.Units = fields[7] }
		if len(fields) > 8 { rec.DeveloperProceeds = fields[8] }
		if len(fields) > 9 { rec.BeginDate = fields[9] }
		if len(fields) > 10 { rec.EndDate = fields[10] }
		if len(fields) > 11 { rec.CustomerCurrency = fields[11] }
		if len(fields) > 12 { rec.CountryCode = fields[12] }
		if len(fields) > 13 { rec.CurrencyProceeds = fields[13] }
		if len(fields) > 14 { rec.AppleIdentifier = fields[14] }
		if len(fields) > 15 { rec.CustomerPrice = fields[15] }
		if len(fields) > 16 { rec.PromoCode = fields[16] }
		if len(fields) > 17 { rec.ParentIdentifier = fields[17] }
		if len(fields) > 18 { rec.Subscription = fields[18] }
		if len(fields) > 19 { rec.Period = fields[19] }
		if len(fields) > 20 { rec.Category = fields[20] }
		if len(fields) > 21 { rec.CMB = fields[21] }
		if len(fields) > 22 { rec.Device = fields[22] }
		if len(fields) > 23 { rec.SupportedPlatforms = fields[23] }
		if len(fields) > 24 { rec.ProceedsReason = fields[24] }
		if len(fields) > 25 { rec.PreservedPricing = fields[25] }
		if len(fields) > 26 { rec.Client = fields[26] }
		if len(fields) > 27 { rec.OrderType = fields[27] }
		records = append(records, rec)
	}

	return records, nil
}
