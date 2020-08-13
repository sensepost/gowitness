package cmd

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sensepost/gowitness/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
)

// handler is the HTTP handler for the web service this command exposes
func handler(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		log.Error("missing url argument")
		return
	}
	log.WithFields(log.Fields{"raw-url": rawURL}).Info("taking screenshot of url")

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		log.WithError(err).Error("Error parsing URL")
		return
	}

	// Generate a safe filename to use
	fname := utils.SafeFileName(u.String()) + ".png"

	// Get the full path where we will be saving the screenshot to
	dst := filepath.Join(chrome.ScreenshotPath, fname)

	log.WithFields(log.Fields{"url": u, "file-name": fname, "destination": dst}).
		Debug("Generated filename for screenshot")

	// Screenshot the URL
	if err := chrome.ScreenshotURL(u, dst); err != nil {
		log.WithFields(log.Fields{"url": u, "error": err}).
			Error("Chrome process reported an error taking screenshot")
		return
	}

	dat, err := ioutil.ReadFile(dst)
	if err != nil {
		log.WithError(err).Error("Error reading saved screenshot file")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(dat)

	// cleanup the file
	log.WithFields(log.Fields{"url": u, "file-name": fname, "destination": dst}).
		Debug("Removing screenshot file")
	os.Remove(dst)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a webservice that takes screenshots",
	Long: `
Start a webservice that takes screenshots. The server starts its
own webserver, and when invoked with the url query parameter,
instructs the underlying Chrome instance to take a screenshot and
return it as the HTTP response.

Assuming the server is hosted on localhost, an HTTP GET request to
take a screenshot of google.com would be:
	http://localhost:7171/?url=https://www.google.com

For example:

$ gowitness server
$ gowitness server --addr 0.0.0.0:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", handler)
		log.WithField("address", serverAddr).Info("server listening")
		log.Fatal(http.ListenAndServe(serverAddr, nil))
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverAddr, "address", "a", "localhost:7171", "server listening address")
}
