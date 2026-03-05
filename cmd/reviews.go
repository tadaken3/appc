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
	reviewsAppID   string
	reviewsRating  int
	reviewsLimit   int
	reviewsSummary bool
	reviewsCountry string
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
	reviewsCmd.Flags().StringVar(&reviewsCountry, "country", "jp", "Country code for rating lookup (used with --summary)")
	_ = reviewsCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(reviewsCmd)
}

func runReviews(cmd *cobra.Command, args []string) error {
	if err := validation.AppID(reviewsAppID); err != nil {
		return err
	}
	if err := validation.CountryCode(reviewsCountry); err != nil {
		return err
	}

	_, c, err := buildClient()
	if err != nil {
		return err
	}

	params := api.ReviewParams{
		AppID:  reviewsAppID,
		Rating: reviewsRating,
		Limit:  reviewsLimit,
	}

	ctx := cmd.Context()
	reviews, err := api.ListReviews(ctx, c, params)
	if err != nil {
		return err
	}

	if reviewsSummary {
		summary := api.SummarizeReviews(reviews)
		rating, err := api.LookupAppRating(ctx, reviewsAppID, reviewsCountry)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "warning: rating lookup: %v\n", err)
		} else {
			summary.AppRating = rating.Rating
			summary.AppRatingCount = rating.RatingCount
		}
		return output.Write(os.Stdout, format, []api.ReviewSummary{summary})
	}

	if len(reviews) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "no reviews found")
		return nil
	}

	return output.Write(os.Stdout, format, reviews)
}
