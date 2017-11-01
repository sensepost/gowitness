package cmd

import (
	"bufio"
	"net/url"
	"os"
	"time"

	"github.com/remeh/sizedwaitgroup" // <3
	chrm "github.com/sensepost/gowitness/chrome"
	"github.com/sensepost/gowitness/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Screenshot URLs sourced from a file",
	Long: `
Screenshot URLs sourced from a file. URLs in the source file should be
newline seperated. Invalid URLs are simply logged and ignored.

For Example:

$ gowitness file -s ~/Desktop/urls
$ gowitness file --source ~/Desktop/urls --threads -2
`,
	Run: func(cmd *cobra.Command, args []string) {

		log.WithField("source", sourceFile).Debug("Reading source file")

		// process the source file
		file, err := os.Open(sourceFile)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "source": sourceFile}).Fatal("Unable to read source file")
		}

		// close the file when we are done with it
		defer file.Close()

		// read each line and populate the channel used to
		// start screenshotting
		scanner := bufio.NewScanner(file)

		swg := sizedwaitgroup.New(maxThreads)
		chrome := chrm.InitChrome()

		for scanner.Scan() {

			candidate := scanner.Text()

			u, err := url.ParseRequestURI(candidate)
			if err != nil {

				log.WithField("url", candidate).Warn("Skipping Invalid URL")
				continue
			}

			swg.Add()

			// Goroutine to run the URL processor
			go func(url *url.URL) {

				defer swg.Done()

				utils.ProcessURL(url, &chrome, waitTimeout)

			}(u)
		}

		swg.Wait()
		log.WithFields(log.Fields{"run-time": time.Since(startTime)}).Info("Complete")

	},
}

func init() {
	RootCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&sourceFile, "source", "s", "", "The source file containing urls")
	fileCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
}
