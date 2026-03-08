package cmd

import (
	"fmt"
	"os"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/spf13/cobra"
)

var (
	subsAppID   string
	subsGroupID string
)

var subscriptionsCmd = &cobra.Command{
	Use:   "subscriptions --app <APP_ID> [--group <GROUP_ID>]",
	Short: "List subscription groups and plans",
	Long:  "Retrieve subscription groups and individual subscription plans for a specific app.\nOutputs JSON by default; use --format csv for CSV output.",
	RunE:  runSubscriptions,
}

func init() {
	subscriptionsCmd.Flags().StringVar(&subsAppID, "app", "", "App ID (required)")
	subscriptionsCmd.Flags().StringVar(&subsGroupID, "group", "", "Filter by subscription group ID")
	_ = subscriptionsCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(subscriptionsCmd)
}

func runSubscriptions(cmd *cobra.Command, args []string) error {
	_, c, err := buildClient()
	if err != nil {
		return err
	}

	params := api.SubscriptionParams{
		AppID:   subsAppID,
		GroupID: subsGroupID,
	}

	ctx := cmd.Context()
	subs, err := api.FetchAppSubscriptions(ctx, c, params)
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "no subscriptions found")
		return nil
	}

	return output.Write(os.Stdout, format, subs)
}
