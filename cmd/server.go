package cmd

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a webservice that takes screenshots",
	Long: `Start a webservice that takes screenshots. The server starts its
own webserver, and when invoked with the url query parameter,
instructs the underlying Chrome instance to take a screenshot and
return it as the HTTP response.

Assuming the server is hosted on localhost, an HTTP GET request to
take a screenshot of google.com would be:
	http://localhost:7171/?url=https://www.google.com`,
	Example: `$ gowitness server
$ gowitness server --addr 0.0.0.0:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		http.HandleFunc("/", handler)
		log.Info().Str("address", options.ServerAddr).Msg("server listening")
		if err := http.ListenAndServe(options.ServerAddr, nil); err != nil {
			log.Fatal().Err(err).Msg("webserver failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&options.ServerAddr, "address", "a", "localhost:7171", "server listening address")
}

// handler is the HTTP handler for the web service this command exposes
func handler(w http.ResponseWriter, r *http.Request) {
	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		http.Error(w, "url parameter missing. eg ?url=https://google.com", http.StatusNotAcceptable)
		return
	}

	url, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf, err := chrm.Screenshot(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf)
}
