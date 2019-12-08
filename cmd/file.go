package cmd

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/reconquest/barely"
	log "github.com/sirupsen/logrus"

	"github.com/remeh/sizedwaitgroup" // <3
	"github.com/sensepost/gowitness/utils"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Screenshot URLs sourced from a file",
	Long: `
Screenshot URLs sourced from a file. URLs in the source file should be
newline separated. Invalid URLs are simply logged and ignored.

For Example:

$ gowitness file -s ~/Desktop/urls
$ gowitness file --source ~/Desktop/urls --threads -2
`,
	Run: func(cmd *cobra.Command, args []string) {

		validateFileCmdFlags()

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

		// Prepare the progress bar to use.
		format, err := template.New("status-bar").
			Parse("  > Processing file: {{if .Updated}}{{end}}{{.Done}}")
		if err != nil {
			log.WithField("err", err).Fatal("Unable to prepare progress bar to use.")
		}
		bar := barely.NewStatusBar(format)
		status := &struct {
			Done    int64
			Updated int64
		}{}
		bar.SetStatus(status)
		bar.Render(os.Stdout)

		for scanner.Scan() {

			candidate := scanner.Text()

			if !(strings.HasPrefix(candidate, `http://`) || strings.HasPrefix(`https://`, candidate)) && (prefixHTTP || prefixHTTPS) {
				if prefixHTTP {
					log.WithFields(log.Fields{"candidate": candidate}).Warn("Prefixing candiate with http://")
					candidate = fmt.Sprintf(`%s%s`, `http://`, candidate)
				} else if prefixHTTPS {
					log.WithFields(log.Fields{"candidate": candidate}).Warn("Prefixing candiate with https://")
					candidate = fmt.Sprintf(`%s%s`, `https://`, candidate)
				} // TODO: Refactor this a bit to support adding both
			}

			u, err := url.ParseRequestURI(candidate)
			if err != nil {

				log.WithField("url", candidate).Warn("Skipping Invalid URL")
				continue
			}

			swg.Add()

			// Goroutine to run the URL processor
			go func(url *url.URL) {

				defer swg.Done()

				utils.ProcessURL(url, &chrome, &db, waitTimeout)

				// update the progress bar
				atomic.AddInt64(&status.Done, 1)
				atomic.AddInt64(&status.Updated, 1)
				bar.Render(os.Stdout)

			}(u)
		}

		swg.Wait()
		bar.Clear(os.Stdout)

		log.WithFields(log.Fields{"run-time": time.Since(startTime)}).Info("Complete")

	},
}

// Validates that the arguments received for fileCmd is valid.
func validateFileCmdFlags() {

	if prefixHTTP && prefixHTTPS {
		log.WithFields(log.Fields{"prefix-http": prefixHTTP, "prefix-https": prefixHTTPS}).
			Fatal("Both --prefix-http and --prefix-https cannot be set")
	}
}

func init() {
	RootCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&sourceFile, "source", "s", "", "The source file containing urls")
	fileCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
	fileCmd.Flags().BoolVarP(&prefixHTTP, "prefix-http", "", false, "Prefix file entries with http:// that have none")
	fileCmd.Flags().BoolVarP(&prefixHTTPS, "prefix-https", "", false, "Prefix file entries with https:// that have none")
}
