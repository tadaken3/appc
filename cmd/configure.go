package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kenta-tanaka/appc/internal/config"
	"github.com/spf13/cobra"
)

var configureValidate bool

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up authentication credentials",
	Long:  "Interactively configure App Store Connect API credentials (Issuer ID, Key ID, private key path, vendor number).",
	RunE:  runConfigure,
}

func init() {
	configureCmd.Flags().BoolVar(&configureValidate, "validate", false, "Validate the current configuration")
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	if configureValidate {
		return runValidate(cmd.ErrOrStderr())
	}
	return runConfigureWith(os.Stdin, cmd.OutOrStdout())
}

func runValidate(out io.Writer) error {
	if err := config.ValidateFromPath(config.DefaultConfigPath()); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "Configuration is valid.")
	return nil
}

func runConfigureWith(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)

	cfgPath := config.DefaultConfigPath()

	existing, _ := config.Load(cfgPath)
	if existing == nil {
		existing = &config.Config{}
	}

	cfg := &config.Config{}

	cfg.IssuerID = prompt(reader, out, "Issuer ID", existing.IssuerID)
	cfg.KeyID = prompt(reader, out, "Key ID", existing.KeyID)
	cfg.PrivateKeyPath = prompt(reader, out, "Private Key Path (.p8 file)", existing.PrivateKeyPath)
	cfg.VendorNumber = prompt(reader, out, "Vendor Number", existing.VendorNumber)

	if err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	_, _ = fmt.Fprintf(out, "\nConfiguration saved to %s\n", cfgPath)
	return nil
}

func prompt(reader *bufio.Reader, out io.Writer, label, current string) string {
	if current != "" {
		_, _ = fmt.Fprintf(out, "%s [%s]: ", label, current)
	} else {
		_, _ = fmt.Fprintf(out, "%s: ", label)
	}
	line, err := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if err != nil || line == "" {
		return current
	}
	return line
}
