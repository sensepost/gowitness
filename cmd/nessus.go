package cmd

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"os"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sensepost/gowitness/lib"
	"github.com/spf13/cobra"
)

// nessusCmd represents the nessus command
var nessusCmd = &cobra.Command{
	Use:   "nessus",
	Short: "Screenshot services from a Nessus XML file",
	Long: `Screenshot services from a Nessus XML file.

To start, export the Nessus results as a .Nessus XML file from the console.

By default, this parser will search for the following match:
	Plugin Name Contains: "Service Detection"

Then it will attempt to identify the web server by:
	Plugin Service Name Contains: "www","https"
	OR
	Plugin Output Value Contains: "web server"

This parser needs a default plugin title to search for. Running this 
command without specifying any --nessus-plugin-contains flags means it
will automatically attempt to find the 'Service Detection' plugin. 
This default plugin appears to be the best plugin for web servers.

You can you can specify --port (multiple times) to only scan specific ports.
If you scan by ports, you still need to use the default --nessus-plugin-contains
flag (or override it with your own value) to identify a plugin to retrieve data
from.

Additionally, you can adjust the --nessus-plugin-output value to search the
plugin output for additional text to search through. The default value is 
'web server'.

You can also adjust the --nessus-service value to include additional service
descriptors. The default values are 'www' and 'https', but perhaps using 
'tcp' could be useful if nessus failed to identify a web server.

Optionally, you may choose to scan the FQDN hostnames with --scan-hostnames.
This will include both IP address and hostnames into the target list.`,
	Example: `
$ gowitness nessus --file output.nessus
$ gowitness nessus --file output.nessus --scan-hostnames

# These options filter services from the nessus file
$ gowitness nessus --file output.nessus --nessus-plugin-output server
$ gowitness nessus --file output.nessus --nessus-service www --nessus-service tcp --nessus-service https
$ gowitness nessus --file output.nessus --no-http
$ gowitness nessus --file output.nessus --no-http --port 8888
$ gowitness nessus --file output.nessus --no-https
$ gowitness nessus --file output.nessus --port 80 --port 8080`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		// prepare targets
		targets, err := getNessusURLs()
		if err != nil {
			log.Fatal().Err(err).Msg("could not process nessus .nessus xml file")
		}
		log.Debug().Int("targets", len(targets)).Msg("number of unique targets")

		// screeny path
		if err = options.PrepareScreenshotPath(); err != nil {
			log.Fatal().Err(err).Msg("failed to prepare the screenshot path")
		}

		// parse headers
		chrm.PrepareHeaderMap()

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
			u, err := url.Parse(target)
			if err != nil {
				log.Warn().Str("url", u.String()).Msg("skipping invalid url")
				continue
			}

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
					log.Debug().Err(err).Str("url", url.String()).Msg("failed to witness url")
				}
			}(u)
		}

		swg.Wait()
		log.Info().Msg("processing complete")
	},
}

func init() {
	rootCmd.AddCommand(nessusCmd)

	nessusCmd.Flags().StringVarP(&options.File, "file", "f", "", "Nessus .nessus XML file")
	nessusCmd.Flags().StringSliceVar(&options.NessusServiceNames, "nessus-service", []string{"www", "https"}, "service name contains filter. supports multiple --service flags")
	nessusCmd.Flags().StringSliceVar(&options.NessusPluginOutput, "nessus-plugin-output", []string{"web server"}, "nessus plugin output contains filter. supports multiple --pluginoutput flags")
	nessusCmd.Flags().StringSliceVar(&options.NessusPluginContains, "nessus-plugin-contains", []string{"Service Detection"}, "nessus plugin name contains filer. supports multiple --plugin-contains flags")
	nessusCmd.Flags().IntSliceVar(&options.NessusPorts, "port", []int{}, "ports filter. supports multiple --port flags")
	nessusCmd.Flags().BoolVarP(&options.NmapScanHostnames, "scan-hostnames", "N", false, "scan hostnames (useful for virtual hosting)")
	nessusCmd.Flags().BoolVarP(&options.NoHTTP, "no-http", "s", false, "do not try using http://")
	nessusCmd.Flags().BoolVarP(&options.NoHTTPS, "no-https", "S", false, "do not try using https://")
	nessusCmd.Flags().IntVarP(&options.Threads, "threads", "t", 4, "threads used to run")

	cobra.MarkFlagRequired(nessusCmd.Flags(), "file")
}

// structure for XML parsing
type reportHost struct {
	HostName    string       `xml:"name,attr"`
	ReportItems []reportItem `xml:"ReportItem"`
	Tags        []tag        `xml:"HostProperties>tag"`
}

type tag struct {
	Key   string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type reportItem struct {
	PluginName   string `xml:"pluginName,attr"`
	ServiceName  string `xml:"svc_name,attr"`
	Port         int    `xml:"port,attr"`
	PluginOutput string `xml:"plugin_output"`
}

// getNessusURLs generates url's from a nessus .nessus xml file based on options
// this function considers many of the flag combinations
func getNessusURLs() (urls []string, err error) {
	log := options.Logger

	// using os.open due to large files
	nessusFile, err := os.Open(options.File)
	if err != nil {
		return
	}
	log.Debug().Str("file", options.File).Msg("reading file")

	defer nessusFile.Close()

	decoder := xml.NewDecoder(nessusFile)

	// Unique maps to cut down on dupliation within nessus files
	var nessusIPsMap = make(map[string][]int)
	var nessusHostsMap = make(map[string][]int)

	for {
		token, err := decoder.Token()

		if err != nil {
			break
		}

		if token == nil {
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			tagName := element.Name.Local

			// Read the ReportHosts from the XML
			if tagName == "ReportHost" {

				var host reportHost
				decoder.DecodeElement(&host, &element)

				// This could be replaced with a map for a quicker retrieval in the future
				// pulling from the tags is a bit annoying
				var fqdn, ip string
				for _, v := range host.Tags {
					if v.Key == "host-fqdn" {
						fqdn = v.Value
					}
					if v.Key == "host-ip" {
						ip = v.Value
					}
				}

				// iterate across the ReportItems XML
				for _, item := range host.ReportItems {

					// leaving this debugging here in case the parser needs to be debugged.
					// dont think this is useful in normal cases

					// log.Debug().Str("IP,Port", ip+" | "+fqdn).Msg("ReportItem: ")
					// log.Debug().Str("Service,PluginName", item.PluginName+" | "+item.ServiceName).Msg("Details: ")

					// skip port if the port does not match the provided ports to filter
					if len(options.NessusPorts) > 0 && !lib.SliceContainsInt(options.NessusPorts, item.Port) {
						continue
					}

					// check the plugin name contains a given string. Contains should work, though startsWith may be useful.
					// A valid plugin name must be given here, otherwise we'll be iterating across too many pointless plugins
					if !lib.SliceContainsString(options.NessusPluginContains, item.PluginName) {
						continue
					}

					// identify that the service is a web server
					if lib.SliceContainsString(options.NessusServiceNames, item.ServiceName) || lib.SliceContainsString(options.NessusPluginOutput, item.PluginOutput) {
						// add the hostnames if the option has been set
						if options.NmapScanHostnames {
							if fqdn != "" {
								nessusHostsMap[fqdn] = removeDuplicatedPorts(append(nessusHostsMap[fqdn], item.Port))
							}
						}
						// checking for empty ip. It should always be set, but you never know
						if ip != "" {
							nessusIPsMap[ip] = removeDuplicatedPorts(append(nessusIPsMap[ip], item.Port))
						}

						log.Debug().Str("target", ip).Str("service", item.ServiceName).Int("port", item.Port).
							Msg("adding target")
					}
				}
			}
		}
	}

	// Build the URL list for unique IPs and Hostnames
	for k, v := range nessusIPsMap {
		urls = append(urls, buildURL(k, v)...)
	}

	for k, v := range nessusHostsMap {
		urls = append(urls, buildURL(k, v)...)
	}

	return
}

// buildURI will build urls taking the http/https options int account
func buildURL(hostname string, port []int) (r []string) {

	for _, v := range port {
		if !options.NoHTTP {
			r = append(r, fmt.Sprintf(`http://%s:%d`, hostname, v))
		}

		if !options.NoHTTPS {
			r = append(r, fmt.Sprintf(`https://%s:%d`, hostname, v))
		}
	}

	return r
}

// removeDuplicatedPort removes duplicated ports
func removeDuplicatedPorts(port []int) []int {
	uniqueMap := make(map[int]bool)
	uniqueList := []int{}

	for _, item := range port {
		if _, ok := uniqueMap[item]; !ok {
			uniqueMap[item] = true
			uniqueList = append(uniqueList, item)
		}
	}

	return uniqueList
}
