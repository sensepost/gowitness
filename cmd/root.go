package cmd

import (
	"os"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/options"
	"github.com/spf13/cobra"
)

var (
	// opts are options set by the command line
	opts = &options.Options{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gowitness",
	Short: "A web screenshot and information gathering tool",
	Long: `               _ _                   
 ___ ___ _ _ _|_| |_ ___ ___ ___ ___ 
| . | . | | | | |  _|   | -_|_ -|_ -|
|_  |___|_____|_|_| |_|_|___|___|___|
|___|    v3, with <3 by @leonjza`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if opts.Logging.Silence {
			log.EnableSilence()
		}

		if opts.Logging.Debug && !opts.Logging.Silence {
			log.EnableDebug()
			log.Debug("debug logging enabled")
		}

		return nil
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// logging configuration
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Debug, "debug-logging", "D", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&opts.Logging.Silence, "silence-logging", "", false, "Silence all (well almost all) logging")
}
