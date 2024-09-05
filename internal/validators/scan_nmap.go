package validators

import (
	"errors"

	"github.com/spf13/cobra"
)

func ValidateScanNmapCmd(cmd *cobra.Command) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	if file == "" {
		return errors.New("a source file name must be set")
	}

	return nil
}
