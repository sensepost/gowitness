package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joeguo/tldextract"
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Use gowitness plugins",
	Long:  ascii.LogoHelp(`Use gowitness plugins`),
}

type NetworkLogResult struct {
	URL string
	Log string
}

func GenerateParentPaths(rawURL string) ([]string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Get the path without query/fragment
	p := parsed.Path
	// Ensure it ends without a slash for consistent splitting
	p = strings.TrimSuffix(p, "/")

	segments := strings.Split(p, "/")
	var results []string

	// Build progressively shorter paths (but not the file itself)
	for i := len(segments) - 1; i > 0; i-- {
		joined := strings.Join(segments[:i], "/") + "/"
		u := *parsed
		u.Path = joined
		results = append(results, strings.Split(u.String(), "?")[0])
	}

	return results, nil
}

var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Extract unique in scope paths for all visited hosts",
	Long:  ascii.LogoHelp(`Take all the visited hosts, loop through loaded resources, pick the ones in scope, enumerate over the possible paths`),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := database.Connection(opts.Writer.DbURI, true, false)
		if err != nil {
			log.Fatal("failed to connect to database", "error", err)
		}
		var result []NetworkLogResult
		c.Raw("SELECT DISTINCT r.url as URL, nl.url as Log FROM results r JOIN network_logs nl ON r.id = nl.result_id").Scan(&result)
		tldExtractor, _ := tldextract.New("tld-cache.txt", false)
		tldCache := map[string]string{}
		for _, r := range result {
			// Extract the registered domain (eTLD+1)
			root, ok := tldCache[r.URL]
			if !ok {
				e := tldExtractor.Extract(r.URL)
				tldCache[r.URL] = e.Root + e.Tld
			}
			l := tldExtractor.Extract(r.Log)
			if l.Root+l.Tld != root {
				continue
			}
			p, _ := GenerateParentPaths(r.Log)
			for _, pp := range p {
				fmt.Println(pp)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.PersistentFlags().StringVar(&opts.Writer.DbURI, "write-db-uri", "sqlite://gowitness.sqlite3", "The database URI to use. Supports SQLite, Postgres, and MySQL (e.g., postgres://user:pass@host:port/db)")
	pluginCmd.AddCommand(pathsCmd)

}
