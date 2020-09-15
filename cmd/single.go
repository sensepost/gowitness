package cmd

import (
	"io/ioutil"
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
		url, err := url.ParseRequestURI(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse target uri")
		}

		// prepare db
		db, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get a db handle")
		}

		var (
			fn string
			fp string
		)
		if options.ScreenshotFileName != "" {
			fn = options.ScreenshotFileName
			fp = lib.ScreenshotPath(options.ScreenshotFileName, url, options.ScreenshotPath)
		} else {
			fn = lib.SafeFileName(url.String())
			fp = lib.ScreenshotPath(fn, url, options.ScreenshotPath)
		}

		log.Debug().Str("url", url.String()).Msg("preflighting")
		resp, title, err := chrm.Preflight(url)
		if err != nil {
			log.Err(err).Msg("preflight failed for url")
			return
		}
		log.Info().Str("url", url.String()).Int("statuscode", resp.StatusCode).Str("title", title).
			Msg("preflight result")

		if db != nil {
			log.Debug().Str("url", url.String()).Msg("storing preflight data")
			if err = chrm.StorePreflight(url, db, resp, title, fn); err != nil {
				log.Error().Err(err).Msg("failed to store preflight information")
			}
		}

		log.Debug().Str("url", url.String()).Msg("screenshotting")
		buf, err := chrm.Screenshot(url)
		if err != nil {
			log.Error().Err(err).Msg("failed to take screenshot")
		}

		log.Debug().Str("url", url.String()).Str("path", fn).Msg("saving screenshot buffer")
		if err := ioutil.WriteFile(fp, buf, 0644); err != nil {
			log.Error().Err(err).Msg("failed to save screenshot buffer")
		}

	},
}

func init() {
	rootCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&options.ScreenshotFileName, "output", "o", "", "write the screenshot to this file")
}
