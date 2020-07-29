package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gosimple/slug"
	chrm "github.com/sensepost/gowitness/chrome"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	serverAddr          string
	serverResolution    string
	serverCacheDir 	    string
	serverChromeTimeout int
)

func handler(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("screenshot of url: ", URL)

	chrome := &chrm.Chrome{
		Resolution:    serverResolution,
		ChromeTimeout: 30,
	}
	chrome.Setup()

	u, err := url.ParseRequestURI(URL)
	if err != nil {
		log.Println("ParseRequestURI error: ", err)
		return
	}
	outputPath := filepath.Join(serverCacheDir, fmt.Sprintf("%s.png", slug.Make(URL)))
	chrome.ScreenshotURL(u, outputPath)

	dat, err := ioutil.ReadFile(outputPath)
	if err != nil {
		log.Println("ReadFile error: ", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(dat)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Take a screenshot of URL with a webservice",
	Long: `
Takes a screenshot of a single given URL and return the image.

For example:

$ gowitness server
$ gowitness server --addr 0.0.0.0:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", handler)
		log.Println("listening on", serverAddr)
		log.Fatal(http.ListenAndServe(serverAddr, nil))
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&serverChromeTimeout, "timeout", "t", 120, "Chrome timeout value.")
        serverCmd.Flags().StringVarP(&serverCacheDir, "cache-dir", "c", "./data", "screenshots cache directory")
	serverCmd.Flags().StringVarP(&serverResolution, "resolution", "r", "800x600", "Screenshot resolution (WidthxHeight)")
	serverCmd.Flags().StringVarP(&serverAddr, "addr", "a", ":7171", "server listening address")
}
