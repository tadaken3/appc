package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
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
		dates, err := dateRange(salesFrom, salesTo)
		if err != nil {
			return err
		}
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
			records, err := api.GetSalesReport(context.Background(), c, params)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "warning: %s: %v\n", date, err)
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

	records, err := api.GetSalesReport(context.Background(), c, params)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, format, records)
}

func dateRange(from, to string) ([]string, error) {
	const layout = "2006-01-02"
	start, err := time.Parse(layout, from)
	if err != nil {
		return nil, fmt.Errorf("invalid --from date %q: use YYYY-MM-DD", from)
	}
	end, err := time.Parse(layout, to)
	if err != nil {
		return nil, fmt.Errorf("invalid --to date %q: use YYYY-MM-DD", to)
	}
	if end.Before(start) {
		return nil, fmt.Errorf("--to must be after --from")
	}
	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(layout))
	}
	return dates, nil
}
