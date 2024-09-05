package cmd

import (
	"github.com/sensepost/gowitness/internal/validators"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "Scan a single target",
	Long:  `Scan a single target`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validators.ValidateScanSingleCmd(cmd); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		runner, err := runner.New(*opts, scanWriters)
		if err != nil {
			log.Error("could not get a runner", "err", err)
			return
		}
		defer runner.Close()

		go func() {
			runner.Targets <- url
			close(runner.Targets)
		}()

		runner.Run()
	},
}

func init() {
	scanCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringP("url", "u", "", "The target to screenshot")
}
