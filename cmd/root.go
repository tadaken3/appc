package cmd

import (
	"github.com/spf13/cobra"
)

var format string
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "appc",
	Short:   "App Store Connect CLI tool",
	Long:    "A CLI tool for fetching App Store Connect data (sales reports, reviews, analytics).",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format (json or csv)")
}

func Execute() error {
	return rootCmd.Execute()
}
