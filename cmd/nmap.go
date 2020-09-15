package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sensepost/gowitness/lib"
	"github.com/spf13/cobra"
	"github.com/tomsteele/go-nmap"
)

// nmapCmd represents the nmap command
var nmapCmd = &cobra.Command{
	Use:   "nmap",
	Short: "Screenshot services from an Nmap XML file",
	Long: `Screenshot services from an Nmap XML file.

When performing an Nmap scan, specify the -oX nmap.xml flag to store
data in an XML formatted file that gowitness can parse.

Running this command without specifying any --services flags means it
will try and screenshot all ports (incl. silly things like SSH etc.).
For this reason, you probably want to rather specify services to probe.
This can be done with the --services / -n flags. For more example
service names parse your local nmap-services file.

For most http-based services, try:
-n http -n http-alt -n http-mgmt -n http-proxy -n https -n https-alt

Alternatively, you can specify --port (multiple times) to only scan
specific ports for hosts. This may be used in conjunction with the
--services flag.

It is also possible to filter for services containing a specific string
with the --service-contains / -w flag. Specifying -w flag as http means
it would match services like http-alt, http-proxy etc.`,
	Example: `# WARNING: These scan all exposed service, like SSH
$ gowitness nmap --nmap-file nmap.xml
$ gowitness nmap --nmap-file nmap.xml --scan-hostnames

# These filter services from the nmap file
$ gowitness nmap --file nmap.xml --service http --service https
$ gowitness nmap --file nmap.xml --service-contains http --service ftp
$ gowitness nmap --file nmap.xml -w http
$ gowitness nmap -f nmap.xml --no-http
$ gowitness nmap -f nmap.xml --no-http --service https --port 8888
$ gowitness nmap -f nmap.xml --no-https -n http -n http-alt
$ gowitness nmap -f nmap.xml --port 80 --port 8080
$ gowitness nmap --nmap-file nmap.xml -s -n http`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		// prepare targets
		targets, err := getNmapURLs()
		if err != nil {
			log.Fatal().Err(err).Msg("could not process nmap xml file")
		}
		log.Debug().Int("targets", len(targets)).Msg("number of targets")

		// screeny path
		if err = options.PrepareScreenshotPath(); err != nil {
			log.Fatal().Err(err).Msg("failed to prepare the screenshot path")
		}

		// prepare db
		db, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get a db handle")
		}

		// prepare swg
		log.Debug().Int("threads", options.Threads).Msg("thread count to use with goroutines")
		swg := sizedwaitgroup.New(options.Threads)

		// process!
		for _, target := range targets {
			u, err := url.ParseRequestURI(target)
			if err != nil {
				log.Warn().Str("url", u.String()).Msg("skipping invalid url")
				continue
			}

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
					if err = chrm.StorePreflight(url, db, resp, title, fn); err != nil {
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

		swg.Wait()
		log.Info().Msg("processing complete")
	},
}

func init() {
	rootCmd.AddCommand(nmapCmd)

	nmapCmd.Flags().StringVarP(&options.NmapFile, "file", "f", "", "nmap xml file")
	nmapCmd.Flags().StringSliceVarP(&options.NmapService, "service", "n", []string{}, "map service name filter. supports multiple --service flags")
	nmapCmd.Flags().StringVarP(&options.NmapServiceContains, "service-contains", "w", "", "partial service name filter (aka: contains)")
	nmapCmd.Flags().IntSliceVarP(&options.NmapPorts, "port", "p", []int{}, "ports filter. supports multiple --port flags")
	nmapCmd.Flags().BoolVarP(&options.NmapScanHostanmes, "scan-hostnames", "N", false, "scan hostnames (useful for virtual hosting)")
	nmapCmd.Flags().BoolVarP(&options.NoHTTP, "no-http", "s", false, "do not try using http://")
	nmapCmd.Flags().BoolVarP(&options.NoHTTPS, "no-https", "S", false, "do not try using https://")
	nmapCmd.Flags().BoolVarP(&options.NmapOpenPortsOnly, "open", "", false, "only select open ports")
	nmapCmd.Flags().IntVarP(&options.Threads, "threads", "t", 4, "threads used to run")

	cobra.MarkFlagRequired(nmapCmd.Flags(), "file")
}

// getNmapURLs generates url's from an nmap xml file based on options
// this function considers many of the flag combinations
func getNmapURLs() (urls []string, err error) {

	xml, err := ioutil.ReadFile(options.NmapFile)
	if err != nil {
		return
	}

	nmapXML, err := nmap.Parse(xml)
	if err != nil {
		return
	}

	// parse the data and generate URL's
	for _, host := range nmapXML.Hosts {
		for _, address := range host.Addresses {
			for _, port := range host.Ports {

				// if we need to filter by service or port, do that
				if len(options.NmapService) > 0 ||
					len(options.NmapPorts) > 0 ||
					len(options.NmapServiceContains) > 0 ||
					options.NmapOpenPortsOnly {

					if lib.SliceContainsString(options.NmapService, port.Service.Name) ||
						(len(options.NmapServiceContains) > 0 &&
							strings.Contains(port.Service.Name, options.NmapServiceContains)) {

						for _, r := range buildURI(address.Addr, port.PortId) {
							urls = append(urls, r)
						}

						if options.NmapScanHostanmes {
							for _, hn := range host.Hostnames {
								for _, r := range buildURI(hn.Name, port.PortId) {
									urls = append(urls, r)
								}
							}
						}
					}

					// add the port if it should be included
					if lib.SliceContainsInt(options.NmapPorts, port.PortId) || (options.NmapOpenPortsOnly && port.State.State == "open") {
						for _, r := range buildURI(address.Addr, port.PortId) {
							urls = append(urls, r)
						}
					}

					// Stop processing, we are filtering by service names
					continue
				}

				// process this without any service name filters
				for _, r := range buildURI(address.Addr, port.PortId) {
					urls = append(urls, r)
				}

				if options.NmapScanHostanmes {
					for _, hn := range host.Hostnames {
						for _, r := range buildURI(hn.Name, port.PortId) {
							urls = append(urls, r)
						}
					}
				}
			}
		}
	}

	return
}

// buildURI will build urls taking the http/https options int account
func buildURI(hostname string, port int) (r []string) {

	if !options.NoHTTP {
		r = append(r, fmt.Sprintf(`http://%s:%d`, hostname, port))
	}

	if !options.NoHTTPS {
		r = append(r, fmt.Sprintf(`https://%s:%d`, hostname, port))
	}

	return r
}
