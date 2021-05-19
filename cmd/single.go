package cmd

import (
	"net/url"

	"github.com/sensepost/gowitness/lib"
	"github.com/spf13/cobra"
)

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single [URL]",
	Short: "Take a screenshot of a single URL",
	Args:  cobra.ExactArgs(1),
	Long: `Takes a screenshot of a single given URL and saves it to a file.

If no --output is provided, a filename for the screenshot will
be automatically generated based on the given URL. If an absolute
output file path is given, the --destination parameter will be
ignored.`,
	Example: `$ gowitness single https://twitter.com
$ gowitness single --destination ~/tweeps_dir https://twitter.com
$ gowitness --disable-db single --destination ~/tweeps_dir https://twitter.com
$ gowitness single -o /screenshots/twitter.png https://twitter.com
$ gowitness single --destination ~/screenshots -o twitter.png https://twitter.com`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		// prepare target
		url, err := url.Parse(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse target uri")
		}

		// prepare db
		db, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get a db handle")
		}

		if err = options.PrepareScreenshotPath(); err != nil {
			log.Fatal().Err(err).Msg("failed to prepare the screenshot path")
		}

		p := &lib.Processor{
			Logger:             log,
			Db:                 db,
			Chrome:             chrm,
			URL:                url,
			ScreenshotPath:     options.ScreenshotPath,
			ScreenshotFileName: options.ScreenshotFileName,
		}

		if err := p.Gowitness(); err != nil {
			log.Debug().Err(err).Str("url", url.String()).Msg("failed to witness url")
		}
	},
}

func init() {
	rootCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&options.ScreenshotFileName, "output", "o", "", "write the screenshot to this file")
}
