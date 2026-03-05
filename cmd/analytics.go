package cmd

import (
	"fmt"
	"os"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/kenta-tanaka/appc/internal/validation"
	"github.com/spf13/cobra"
)

var (
	analyticsAppID    string
	analyticsCategory string
	analyticsSnapshot bool
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
  PERFORMANCE            App performance metrics (crashes, disk writes, etc.)

Use --snapshot to retrieve historical data (from app creation to request date).
Without --snapshot, only data from the ONGOING request creation date onward is available.`,
	RunE: runAnalytics,
}

func init() {
	analyticsCmd.Flags().StringVar(&analyticsAppID, "app", "", "App ID (required)")
	analyticsCmd.Flags().StringVar(&analyticsCategory, "category", "",
		"Report category (APP_USAGE, APP_STORE_ENGAGEMENT, COMMERCE, FRAMEWORK_USAGE, PERFORMANCE)")
	analyticsCmd.Flags().BoolVar(&analyticsSnapshot, "snapshot", false,
		"Use ONE_TIME_SNAPSHOT to retrieve historical data from app creation date")
	_ = analyticsCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(analyticsCmd)
}

func runAnalytics(cmd *cobra.Command, args []string) error {
	if err := validation.AppID(analyticsAppID); err != nil {
		return err
	}

	_, c, err := buildClient()
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessType := "ONGOING"
	if analyticsSnapshot {
		accessType = "ONE_TIME_SNAPSHOT"
	}

	log := func(format string, a ...any) {
		_, _ = fmt.Fprintf(os.Stderr, format, a...)
	}

	allRecords, err := api.FetchAllSegmentRecords(ctx, c, analyticsAppID, accessType, analyticsCategory, log)
	if err != nil {
		return err
	}

	if len(allRecords) == 0 {
		fmt.Fprintln(os.Stderr, "No segment data available.")
		return output.Write(os.Stdout, format, []api.AnalyticsSegmentRecord{})
	}

	return output.Write(os.Stdout, format, allRecords)
}
