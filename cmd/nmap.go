package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/reconquest/barely"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sensepost/gowitness/utils"
	log "github.com/sirupsen/logrus"
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

For example:

# WARNING: These scan all exposed service, like SSH
$ gowitness nmap --nmap-file nmap.xml
$ gowitness nmap --nmap-file nmap.xml --scan-hostnames

# These filter services from the nmap file
$ gowitness nmap --nmap-file nmap.xml --service http --service https
$ gowitness nmap -f nmap.xml --no-http
$ gowitness nmap -f nmap.xml --no-http --service https --port 8888
$ gowitness nmap -f nmap.xml --no-https -n http -n http-alt
$ gowitness nmap -f nmap.xml --port 80 --port 8080
$ gowitness nmap --nmap-file nmap.xml -s -n http`,
	Run: func(cmd *cobra.Command, args []string) {

		validateNmapFlags()

		log.WithFields(log.Fields{"file": nmapFile}).Info("Parsing nmap file")
		xml, err := ioutil.ReadFile(nmapFile)
		if err != nil {
			log.WithFields(log.Fields{"file": nmapFile, "err": err}).Fatal("Error reading nmap file")
		}

		f, err := nmap.Parse(xml)
		if err != nil {
			log.WithFields(log.Fields{"file": nmapFile, "err": err}).Fatal("Error parsing nmap file")
		}
		log.WithFields(log.Fields{"args": f.Args}).Info("Parsed NMAP file information, generating URL's")

		targets := parseNmapURLs(f)
		log.WithField("target-count", len(targets)).Info("Total targets to be processed")

		// Start processing the calculated permutations
		log.WithField("thread-count", maxThreads).Debug("Maximum threads")
		swg := sizedwaitgroup.New(maxThreads)

		// Prepare the progress bar to use.
		format, err := template.New("status-bar").
			Parse("  > Processing range: {{if .Updated}}{{end}}{{.Done}}/{{.Total}}")
		if err != nil {
			log.WithField("err", err).Fatal("Unable to prepare progress bar to use.")
		}
		bar := barely.NewStatusBar(format)
		status := &struct {
			Total   int
			Done    int64
			Updated int64
		}{
			Total: len(targets),
		}
		bar.SetStatus(status)
		bar.Render(os.Stdout)

		for _, target := range targets {

			u, err := url.ParseRequestURI(target)
			if err != nil {

				log.WithField("url", target).Warn("Skipping Invalid URL")
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

		log.WithFields(log.Fields{"run-time": time.Since(startTime), "target-count": len(targets)}).
			Info("Complete")
	},
}

// parseNmapURLs parses an openeed nmap XML and returns URL's
func parseNmapURLs(nmapXML *nmap.NmapRun) []string {

	var u []string

	// parse the data and generate URL's
	for _, host := range nmapXML.Hosts {
		for _, address := range host.Addresses {
			for _, port := range host.Ports {

				// if we need to filter by service or port, do that
				if len(nmapServices) > 0 || len(nmapPorts) > 0 {
					if utils.SliceContainsString(nmapServices, port.Service.Name) {

						for _, r := range buildURI(address.Addr, port.PortId) {
							u = append(u, r)
						}

						if scanHostnames {
							for _, hn := range host.Hostnames {
								for _, r := range buildURI(hn.Name, port.PortId) {
									u = append(u, r)
								}
							}
						}
					}

					// add the port if it should be included
					if utils.SliceContainsInt(nmapPorts, port.PortId) {
						for _, r := range buildURI(address.Addr, port.PortId) {
							u = append(u, r)
						}
					}

					// Stop processing, we are filtering by service names
					continue
				}

				// process this without any service name filters
				for _, r := range buildURI(address.Addr, port.PortId) {
					u = append(u, r)
				}

				if scanHostnames {
					for _, hn := range host.Hostnames {
						for _, r := range buildURI(hn.Name, port.PortId) {
							u = append(u, r)
						}
					}
				}
			}
		}
	}

	return u
}

func buildURI(hostname string, port int) []string {
	var r []string

	if !skipHTTP {
		r = append(r, fmt.Sprintf(`http://%s:%d`, hostname, port))
	}

	if !skipHTTPS {
		r = append(r, fmt.Sprintf(`https://%s:%d`, hostname, port))
	}

	return r
}

func init() {
	RootCmd.AddCommand(nmapCmd)

	nmapCmd.Flags().StringVarP(&nmapFile, "nmap-file", "f", "", "The source file containing urls")
	nmapCmd.Flags().StringSliceVarP(&nmapServices, "service", "n", []string{}, "Nmap service names to filter by. Multiple --service flags are supported")
	nmapCmd.Flags().IntSliceVarP(&nmapPorts, "port", "p", []int{}, "Nmap ports to filter by. Multiple --port flags are supported")
	nmapCmd.Flags().BoolVarP(&scanHostnames, "scan-hostnames", "N", false, "Also scan hostnames (useful for virtual hosting)")
	nmapCmd.Flags().BoolVarP(&skipHTTP, "no-http", "s", false, "Skip trying to connect with HTTP")
	nmapCmd.Flags().BoolVarP(&skipHTTPS, "no-https", "S", false, "Skip trying to connect with HTTPS")
	nmapCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
	cobra.MarkFlagRequired(nmapCmd.Flags(), "source")
}

func validateNmapFlags() {

	if skipHTTP && skipHTTPS {
		log.Fatal("Cannot disable both http and https scanning")
	}
}
