package cmd

import (
	"fmt"
	"os"

	"github.com/tadaken3/appc/internal/api"
	"github.com/tadaken3/appc/internal/auth"
	"github.com/tadaken3/appc/internal/client"
	"github.com/tadaken3/appc/internal/config"
	"github.com/tadaken3/appc/internal/output"
	"github.com/spf13/cobra"
)

var appsCountry string

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List all apps",
	Long:  "Retrieve a list of all apps in your App Store Connect account.",
	RunE:  runApps,
}

func init() {
	appsCmd.Flags().StringVar(&appsCountry, "country", "jp", "Country code for rating lookup")
	rootCmd.AddCommand(appsCmd)
}

func runApps(cmd *cobra.Command, args []string) error {
	_, c, err := buildClient()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	apps, err := api.ListApps(ctx, c)
	if err != nil {
		return err
	}

	// Enrich with ratings from iTunes Lookup API
	var ids []string
	for _, a := range apps {
		ids = append(ids, a.ID)
	}
	if len(ids) > 0 {
		warn := func(id string, err error) {
			_, _ = fmt.Fprintf(os.Stderr, "warning: rating lookup %s: %v\n", id, err)
		}
		ratings := api.LookupAppRatings(ctx, ids, appsCountry, warn)
		for i, a := range apps {
			if r, ok := ratings[a.ID]; ok {
				apps[i].Rating = r.Rating
				apps[i].RatingCount = r.RatingCount
			}
		}
	}

	return output.Write(os.Stdout, format, apps)
}

func buildClient() (*config.Config, *client.Client, error) {
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.Resolve(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading config (run 'appc configure' first): %w", err)
	}
	if err := config.Validate(cfg); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	keyPEM, err := cfg.PrivateKeyPEM()
	if err != nil {
		return nil, nil, fmt.Errorf("setting up auth: %w", err)
	}

	tokenGen, err := auth.NewTokenGeneratorFromPEM(cfg.IssuerID, cfg.KeyID, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up auth: %w", err)
	}

	return cfg, client.New(tokenGen), nil
}
