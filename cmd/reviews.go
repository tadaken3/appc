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
	reviewsAppID  string
	reviewsRating int
	reviewsLimit  int
	reviewsSummary bool
)

var reviewsCmd = &cobra.Command{
	Use:   "reviews",
	Short: "Fetch customer reviews",
	Long:  "Retrieve customer reviews for a specific app.",
	RunE:  runReviews,
}

func init() {
	reviewsCmd.Flags().StringVar(&reviewsAppID, "app", "", "App ID (required)")
	reviewsCmd.Flags().IntVar(&reviewsRating, "rating", 0, "Filter by rating (1-5)")
	reviewsCmd.Flags().IntVar(&reviewsLimit, "limit", 100, "Maximum number of reviews")
	reviewsCmd.Flags().BoolVar(&reviewsSummary, "summary", false, "Show rating summary instead of full reviews")
	_ = reviewsCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(reviewsCmd)
}

func runReviews(cmd *cobra.Command, args []string) error {
	_, c, err := buildClient()
	if err != nil {
		return err
	}

	params := api.ReviewParams{
		AppID:  reviewsAppID,
		Rating: reviewsRating,
		Limit:  reviewsLimit,
	}

	reviews, err := api.ListReviews(context.Background(), c, params)
	if err != nil {
		return err
	}

	if reviewsSummary {
		summary := api.SummarizeReviews(reviews)
		return output.Write(os.Stdout, format, []api.ReviewSummary{summary})
	}

	if len(reviews) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "no reviews found")
		return nil
	}

	return output.Write(os.Stdout, format, reviews)
}
