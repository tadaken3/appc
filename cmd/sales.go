package cmd

import (
	"context"
	"os"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/spf13/cobra"
)

var (
	salesReportType    string
	salesReportDate    string
	salesFrequency     string
	salesReportSubType string
	salesVendorNumber  string
)

var salesCmd = &cobra.Command{
	Use:   "sales",
	Short: "Fetch sales reports",
	Long:  "Download and parse App Store Connect sales reports (gzip+TSV).",
	RunE:  runSales,
}

func init() {
	salesCmd.Flags().StringVar(&salesReportType, "type", "SALES", "Report type (SALES, PRE_ORDER, NEWSSTAND)")
	salesCmd.Flags().StringVar(&salesReportDate, "date", "", "Report date (YYYY-MM-DD)")
	salesCmd.Flags().StringVar(&salesFrequency, "frequency", "DAILY", "Report frequency (DAILY, WEEKLY, MONTHLY, YEARLY)")
	salesCmd.Flags().StringVar(&salesReportSubType, "sub-type", "SUMMARY", "Report sub type (SUMMARY, DETAILED, OPT_IN)")
	salesCmd.Flags().StringVar(&salesVendorNumber, "vendor", "", "Vendor number (overrides config)")
	salesCmd.MarkFlagRequired("date")
	rootCmd.AddCommand(salesCmd)
}

func runSales(cmd *cobra.Command, args []string) error {
	cfg, c, err := buildClient()
	if err != nil {
		return err
	}

	vendor := salesVendorNumber
	if vendor == "" {
		vendor = cfg.VendorNumber
	}

	params := api.SalesReportParams{
		ReportType:    salesReportType,
		ReportDate:    salesReportDate,
		Frequency:     salesFrequency,
		ReportSubType: salesReportSubType,
		VendorNumber:  vendor,
	}

	records, err := api.GetSalesReport(context.Background(), c, params)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, format, records)
}
