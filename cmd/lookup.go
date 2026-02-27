package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/spf13/cobra"
)

var (
	lookupAppIDs  string
	lookupCountry string
)

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Look up app ratings from iTunes",
	Long:  "Retrieve app ratings and metadata from the iTunes Lookup API. Supports multiple App IDs for competitor research.",
	RunE:  runLookup,
}

func init() {
	lookupCmd.Flags().StringVar(&lookupAppIDs, "app", "", "App ID(s), comma-separated (required)")
	lookupCmd.Flags().StringVar(&lookupCountry, "country", "jp", "Country code for store lookup")
	_ = lookupCmd.MarkFlagRequired("app")
	rootCmd.AddCommand(lookupCmd)
}

func runLookup(cmd *cobra.Command, args []string) error {
	ids := strings.Split(lookupAppIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	ctx := context.Background()
	warn := func(id string, err error) {
		_, _ = fmt.Fprintf(os.Stderr, "warning: lookup %s: %v\n", id, err)
	}

	ratings := api.LookupAppRatings(ctx, ids, lookupCountry, warn)

	var results []api.AppRating
	for _, id := range ids {
		if r, ok := ratings[id]; ok {
			results = append(results, r)
		}
	}

	return output.Write(os.Stdout, format, results)
}
