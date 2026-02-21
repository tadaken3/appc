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
	Long:  "Request and retrieve App Store Connect analytics reports.",
	RunE:  runAnalytics,
}

func init() {
	analyticsCmd.Flags().StringVar(&analyticsAppID, "app", "", "App ID (required)")
	analyticsCmd.Flags().StringVar(&analyticsCategory, "category", "", "Report category (e.g., APP_USAGE, APP_STORE_ENGAGEMENT)")
	analyticsCmd.MarkFlagRequired("app")
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

	return output.Write(os.Stdout, format, reports)
}
