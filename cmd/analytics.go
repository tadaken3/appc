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

Downloads segment CSV data and outputs parsed records in JSON or CSV format.

Available categories:
  APP_USAGE              App usage metrics (sessions, active devices, etc.)
  APP_STORE_ENGAGEMENT   App Store engagement (impressions, page views, etc.)
  COMMERCE               Commerce data (sales, in-app purchases, etc.)
  FRAMEWORK_USAGE        Framework usage data
  PERFORMANCE            App performance metrics (crashes, disk writes, etc.)`,
	RunE: runAnalytics,
}

func init() {
	analyticsCmd.Flags().StringVar(&analyticsAppID, "app", "", "App ID (required)")
	analyticsCmd.Flags().StringVar(&analyticsCategory, "category", "",
		"Report category (APP_USAGE, APP_STORE_ENGAGEMENT, COMMERCE, FRAMEWORK_USAGE, PERFORMANCE)")
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
		return output.Write(os.Stdout, format, reports)
	}

	// Collect segment CSV data: reports → instances → segments → CSV download
	var allRecords []api.AnalyticsSegmentRecord
	for _, report := range reports {
		fmt.Fprintf(os.Stderr, "Fetching instances for report %q (%s)...\n", report.Name, report.Category)

		instances, err := api.GetAnalyticsInstances(ctx, c, report.ID)
		if err != nil {
			return fmt.Errorf("getting instances for report %s: %w", report.ID, err)
		}

		if len(instances) == 0 {
			fmt.Fprintf(os.Stderr, "  No instances found, skipping.\n")
			continue
		}

		// Use the latest instance (most recent processing date)
		latest := instances[len(instances)-1]
		fmt.Fprintf(os.Stderr, "Using instance %s (date: %s)...\n", latest.ID, latest.ProcessingDate)

		segments, err := api.GetAnalyticsSegments(ctx, c, latest.ID)
		if err != nil {
			return fmt.Errorf("getting segments for instance %s: %w", latest.ID, err)
		}

		for _, seg := range segments {
			if seg.URL == "" {
				continue
			}
			fmt.Fprintf(os.Stderr, "Downloading segment %s (%d bytes)...\n", seg.ID, seg.SizeInBytes)

			records, err := api.DownloadAnalyticsSegmentCSV(ctx, seg.URL)
			if err != nil {
				return fmt.Errorf("downloading segment %s: %w", seg.ID, err)
			}
			allRecords = append(allRecords, records...)
		}
	}

	if len(allRecords) == 0 {
		fmt.Fprintln(os.Stderr, "No segment data available.")
		return output.Write(os.Stdout, format, []api.AnalyticsReport{})
	}

	return output.Write(os.Stdout, format, allRecords)
}
