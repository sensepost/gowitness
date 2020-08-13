package cmd

import (
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/sensepost/gowitness/utils"
	"github.com/spf13/cobra"
)

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "Take a screenshot of a single URL",
	Long: `
Takes a screenshot of a single given URL and saves it to a file.
If no --output is provided, a filename for the screenshot will
be automatically generated based on the given URL. If an absolute
output file path is given, the --destination parameter will be
ignored.

For example:

$ gowitness single --url https://twitter.com
$ gowitness single --destination ~/tweeps_dir --url https://twitter.com
$ gowitness single -u https://twitter.com
$ gowitness single -o /screenshots/twitter.png -u https://twitter.com
$ gowitness single --destination ~/screenshots -o twitter.png -u https://twitter.com`,

	Run: func(cmd *cobra.Command, args []string) {

		u, err := url.ParseRequestURI(screenshotURL)
		if err != nil {
			log.WithField("url", screenshotURL).Fatal("Invalid URL specified")
		}

		// Process this URL
		utils.ProcessURL(u, &chrome, &db, waitTimeout, outputFile)

		log.WithFields(log.Fields{"run-time": time.Since(startTime)}).Info("Complete")
	},
}

func init() {
	RootCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&screenshotURL, "url", "u", "", "The URL to screenshot")
	singleCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Write the screenshot to this file")
}
