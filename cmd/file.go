package cmd

import (
	"bufio"
	"net/url"
	"os"
	"strings"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sensepost/gowitness/lib"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file [input]",
	Short: "screenshot URLs sourced from a file or stdin",
	Long: `Screenshot URLs sourced from a file or stdin.

URLs in the source file should be newline separated. Invalid URLs are simply
logged and ignored.`,
	Example: `$ gowitness file -f ~/Desktop/urls
$ gowitness file -f urls.txt --threads 2
$ cat urls.txt | gowitness file -f -
$ gowitness file -f <( shuf domains ) --no-http`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		scanner, f, err := getScanner(options.File)
		if err != nil {
			log.Fatal().Err(err).Str("file", options.File).Msg("unable to read source file")
		}
		defer f.Close()

		db, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get a db handle")
		}

		log.Debug().Int("threads", options.Threads).Msg("thread count to use with goroutines")
		swg := sizedwaitgroup.New(options.Threads)

		if err = options.PrepareScreenshotPath(); err != nil {
			log.Fatal().Err(err).Msg("failed to prepare the screenshot path")
		}

		// parse headers
		chrm.PrepareHeaderMap()

		for scanner.Scan() {
			candidate := scanner.Text()
			if candidate == "" {
				return
			}

			for _, u := range getUrls(candidate) {
				swg.Add()

				log.Debug().Str("url", u.String()).Msg("queueing goroutine for url")
				go func(url *url.URL) {
					defer swg.Done()

					p := &lib.Processor{
						Logger:         log,
						Db:             db,
						Chrome:         chrm,
						URL:            url,
						ScreenshotPath: options.ScreenshotPath,
					}

					if err := p.Gowitness(); err != nil {
						log.Error().Err(err).Str("url", url.String()).Msg("failed to witness url")
					}
				}(u)
			}
		}

		swg.Wait()
		log.Info().Msg("processing complete")
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&options.File, "file", "f", "", "file containing urls. use - for stdin")
	fileCmd.Flags().IntVarP(&options.Threads, "threads", "t", 4, "threads used to run")
	fileCmd.Flags().BoolVar(&options.NoHTTPS, "no-https", false, "do not prefix https:// where missing")
	fileCmd.Flags().BoolVar(&options.NoHTTP, "no-http", false, "do not prefix http:// where missing")

	cobra.MarkFlagRequired(fileCmd.Flags(), "file")
}

// getScanner prepares a bufio.Scanner to read from either
// stdin, or a file.
// the size attribute > 0 will be returned if a file was the input
// it is up to the caller to close the file.
func getScanner(i string) (*bufio.Scanner, *os.File, error) {
	if i == "-" {
		return bufio.NewScanner(os.Stdin), nil, nil
	}

	file, err := os.Open(i)
	if err != nil {
		return nil, nil, err
	}

	return bufio.NewScanner(file), file, nil
}

// getUrls generates urls for an incoming target depending
// on wether the target has an http prefix and the flags set
func getUrls(target string) (c []*url.URL) {

	// if there already is a protocol, just parse and add that
	if strings.HasPrefix(target, "http") {
		u, err := url.Parse(target)
		if err == nil {
			c = append(c, u)
		}

		return
	}

	if !strings.HasPrefix(target, "http://") && !options.NoHTTP {
		u, err := url.Parse("http://" + target)
		if err == nil {
			c = append(c, u)
		}
	}

	if !strings.HasPrefix(target, "https://") && !options.NoHTTPS {
		u, err := url.Parse("https://" + target)
		if err == nil {
			c = append(c, u)
		}
	}

	return
}
