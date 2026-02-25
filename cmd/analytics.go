package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/spf13/cobra"
)

var (
	analyticsAppID    string
	analyticsCategory string
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Fetch analytics data",
	Long: `Request and retrieve App Store Connect analytics reports.

Downloads segment CSV data for the specified app and outputs it.

Examples:
  # Fetch APP_USAGE analytics for an app
  appc analytics --app 123456789 --category APP_USAGE

  # Fetch APP_STORE_ENGAGEMENT data as CSV
  appc analytics --app 123456789 --category APP_STORE_ENGAGEMENT --format csv

  # Fetch all categories (no filter)
  appc analytics --app 123456789`,
	RunE: runAnalytics,
}

func init() {
	analyticsCmd.Flags().StringVar(&analyticsAppID, "app", "", "App ID (required)")
	analyticsCmd.Flags().StringVar(&analyticsCategory, "category", "",
		"Report category filter (e.g., APP_USAGE, APP_STORE_ENGAGEMENT, APP_STORE_DISCOVERY, PERFORMANCE)")
	_ = analyticsCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(analyticsCmd)
}

func runAnalytics(cmd *cobra.Command, args []string) error {
	_, c, err := buildClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	fmt.Fprintln(os.Stderr, "Requesting analytics report...")
	reqID, err := api.RequestAnalyticsReport(ctx, c, analyticsAppID)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Request ID: %s\n", reqID)

	reports, err := api.GetAnalyticsReports(ctx, c, reqID, analyticsCategory)
	if err != nil {
		return err
	}

	if len(reports) == 0 {
		fmt.Fprintln(os.Stderr, "No reports found.")
		return output.Write(os.Stdout, format, []api.AnalyticsDataRecord{})
	}

	var allRecords []api.AnalyticsDataRecord
	for _, report := range reports {
		fmt.Fprintf(os.Stderr, "Fetching segments for report %q (%s)...\n", report.Name, report.Category)

		segments, err := api.GetAnalyticsSegments(ctx, c, report.ID)
		if err != nil {
			return err
		}

		for _, seg := range segments {
			fmt.Fprintf(os.Stderr, "Downloading segment %s...\n", seg.ID)
			records, err := api.DownloadSegmentCSV(ctx, seg.URL)
			if err != nil {
				return fmt.Errorf("downloading segment %s: %w", seg.ID, err)
			}
			allRecords = append(allRecords, records...)
		}
	}

	fmt.Fprintf(os.Stderr, "Total records: %d\n", len(allRecords))
	return output.Write(os.Stdout, format, allRecords)
}
