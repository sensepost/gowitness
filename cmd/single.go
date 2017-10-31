package cmd

import (
	"net/url"
	"strings"
	"time"

	chrm "github.com/leonjza/gowitness/chrome"
	"github.com/leonjza/gowitness/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// singleCmd represents the single command
var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "Take a screenshot of a single URL",
	Long: `
Takes a screenshot of a single given URL and saves it to a file.
If no --destination is provided, a filename for the screenshot will
be automatically generated based on the given URL.

For example:

$ gowitness single --url https://twitter.com
$ gowitness single --destination tweeps_page.png --url https://twitter.com
$ gowitness single -u https://twitter.com`,

	Run: func(cmd *cobra.Command, args []string) {

		u, err := url.ParseRequestURI(screenshotURL)
		if err != nil {

			log.WithField("url", screenshotURL).Fatal("Invalid URL specified")
		}

		if screenshotDestination == "" {
			screenshotDestination = utils.SafeFileName(u.String())
			log.WithField("destination", screenshotDestination).Debug("Generated filename for screenshot")
		}

		if !strings.HasSuffix(screenshotDestination, ".png") {

			log.WithField("destination", screenshotDestination).Debug("Adding .png suffix to destination")
			screenshotDestination = screenshotDestination + ".png"
		}

		chrome := chrm.InitChrome()
		utils.ProcessURL(u, &chrome, waitTimeout)

		log.WithFields(log.Fields{"run-time": time.Since(startTime)}).Info("Complete")
	},
}

func init() {
	RootCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&screenshotURL, "url", "u", "", "The URL to screenshot")
	singleCmd.Flags().StringVarP(&screenshotDestination,
		"destination", "d", "", "The destination filename of the screenshot")
}
