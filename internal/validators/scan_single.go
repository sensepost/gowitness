package validators

import (
	"errors"

	"github.com/spf13/cobra"
)

// ValidateScanSingleCmd validates the scan file subcommand
func ValidateScanSingleCmd(cmd *cobra.Command) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	if url == "" {
		return errors.New("a url must be set")
	}

	return nil
}
