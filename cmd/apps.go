package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kenta-tanaka/appc/internal/api"
	"github.com/kenta-tanaka/appc/internal/auth"
	"github.com/kenta-tanaka/appc/internal/client"
	"github.com/kenta-tanaka/appc/internal/config"
	"github.com/kenta-tanaka/appc/internal/output"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List all apps",
	Long:  "Retrieve a list of all apps in your App Store Connect account.",
	RunE:  runApps,
}

func init() {
	rootCmd.AddCommand(appsCmd)
}

func runApps(cmd *cobra.Command, args []string) error {
	cfg, c, err := buildClient()
	if err != nil {
		return err
	}
	_ = cfg

	apps, err := api.ListApps(context.Background(), c)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, format, apps)
}

func buildClient() (*config.Config, *client.Client, error) {
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading config (run 'appc configure' first): %w", err)
	}

	tokenGen, err := auth.NewTokenGenerator(cfg.IssuerID, cfg.KeyID, cfg.PrivateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up auth: %w", err)
	}

	return cfg, client.New(tokenGen), nil
}
