package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/tadaken3/appc/internal/api"
	"github.com/tadaken3/appc/internal/output"
	"github.com/tadaken3/appc/internal/sales"
	"github.com/spf13/cobra"
)

var (
	salesReportType    string
	salesReportDate    string
	salesFrom          string
	salesTo            string
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
	salesCmd.Flags().StringVar(&salesReportDate, "date", "", "Report date (YYYY-MM-DD or YYYY-MM)")
	salesCmd.Flags().StringVar(&salesFrom, "from", "", "Start date for range (YYYY-MM-DD, requires --to, daily only)")
	salesCmd.Flags().StringVar(&salesTo, "to", "", "End date for range (YYYY-MM-DD, requires --from, daily only)")
	salesCmd.Flags().StringVar(&salesFrequency, "frequency", "DAILY", "Report frequency (DAILY, WEEKLY, MONTHLY, YEARLY)")
	salesCmd.Flags().StringVar(&salesReportSubType, "sub-type", "SUMMARY", "Report sub type (SUMMARY, DETAILED, OPT_IN)")
	salesCmd.Flags().StringVar(&salesVendorNumber, "vendor", "", "Vendor number (overrides config)")
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

	// --from / --to による複数日付一括取得
	if salesFrom != "" || salesTo != "" {
		if salesFrom == "" || salesTo == "" {
			return fmt.Errorf("--from and --to must be used together")
		}
		if strings.ToUpper(salesFrequency) != "DAILY" {
			return fmt.Errorf("--from and --to are supported only with --frequency DAILY")
		}
		dates, err := sales.DateRange(salesFrom, salesTo)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		var all []api.SalesRecord
		for _, date := range dates {
			_, _ = fmt.Fprintf(os.Stderr, "fetching %s...\n", date)
			params := api.SalesReportParams{
				ReportType:    salesReportType,
				ReportDate:    date,
				Frequency:     salesFrequency,
				ReportSubType: salesReportSubType,
				VendorNumber:  vendor,
			}
			records, err := api.GetSalesReport(ctx, c, params)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: %s: %v\n", date, err)
				continue
			}
			if len(records) == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "no sales data for %s\n", date)
				continue
			}
			all = append(all, records...)
		}
		return output.Write(os.Stdout, format, all)
	}

	if salesReportDate == "" {
		return fmt.Errorf("--date is required (or use --from and --to for a range)")
	}

	params := api.SalesReportParams{
		ReportType:    salesReportType,
		ReportDate:    salesReportDate,
		Frequency:     salesFrequency,
		ReportSubType: salesReportSubType,
		VendorNumber:  vendor,
	}

	records, err := api.GetSalesReport(cmd.Context(), c, params)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "no sales data for %s\n", salesReportDate)
	}

	return output.Write(os.Stdout, format, records)
}
