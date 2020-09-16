package cmd

import (
	"bufio"
	"io/ioutil"
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
	Long: `Screenshot URLs sourced from a file or stdin. URLs in the source
file should be newline separated. Invalid URLs are simply logged and ignored.`,
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

					// file name / path
					fn := lib.SafeFileName(url.String())
					fp := lib.ScreenshotPath(fn, url, options.ScreenshotPath)

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
						if _, err := chrm.StorePreflight(url, db, resp, title, fn); err != nil {
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

// getInput determines what the file input should be
// without any file argument we will assume stdin with -
func getInput(a []string) (input string) {
	if len(a) <= 0 {
		input = "-"
	} else {
		input = a[0]
	}
	return
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
		u, err := url.ParseRequestURI(target)
		if err == nil {
			c = append(c, u)
		}

		return
	}

	if !strings.HasPrefix(target, "http://") && !options.NoHTTP {
		u, err := url.ParseRequestURI("http://" + target)
		if err == nil {
			c = append(c, u)
		}
	}

	if !strings.HasPrefix(target, "https://") && !options.NoHTTPS {
		u, err := url.ParseRequestURI("https://" + target)
		if err == nil {
			c = append(c, u)
		}
	}

	return
}
