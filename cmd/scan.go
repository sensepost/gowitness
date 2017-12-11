package cmd

import (
	"net/url"
	"os"
	"sync/atomic"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/reconquest/barely"
	"github.com/remeh/sizedwaitgroup" // <3
	"github.com/sensepost/gowitness/utils"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a CIDR range and take screenshots along the way",
	Long: `
Scans a CIDR range and takes screenshots along the way!
This command takes a CIDR, ports and flag arguments to specify wether
it is nessesary to connect via HTTP and or HTTPS to urls. The
combination of these flags are used to generate permutations that
are iterated over and processed.

When specifying the --random/-r flag, the ip:port permutations that are
generated will go through a shuffling phase so that the resultant
requests that are made wont follow each other on the same host.
This may be useful in cases where too many ports specified by the
--ports flag might trigger port scan alerts.

For example:

$ gowitness scan --cidr 192.168.0.0/24
$ gowitness scan --cidr 192.168.0.0/24 --cidr 10.10.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.1/32 --no-https
$ gowitness --log-level debug scan --threads 20 --ports 80,443,8080 --no-http --cidr 192.168.0.0/30
`,
	Run: func(cmd *cobra.Command, args []string) {

		validateScanCmdFlags()

		ports, _ := utils.Ports(scanPorts)
		log.WithField("ports", ports).Debug("Using ports")

		if len(ports) <= 0 {
			log.WithField("ports", scanPorts).Fatal("Please specify at least one port to connect to")
			return
		}

		var ips []string
		log.WithField("cidr", scanCidr).Debug("Using CIDR ranges")

		// loop and parse the --cidr flags we got
		for _, cidr := range scanCidr {

			// parse the cidr
			cidrIps, err := utils.Hosts(cidr)
			if err != nil {
				log.WithFields(log.Fields{"cidr": scanCidr, "error": err}).Fatal("Failed to parse CIDR")
				return
			}

			// append the ips from the current cidr
			log.WithFields(log.Fields{"cidr": cidr, "cidr-ips": len(cidrIps)}).Debug("Appending cidr")
			ips = append(ips, cidrIps...)
		}

		log.WithFields(log.Fields{"total-ips": len(ips)}).Debug("Finished parsing CIDR ranges")

		permutations, err := utils.Permutations(ips, ports, skipHTTP, skipHTTPS)

		if randomPermutations {
			log.WithFields(log.Fields{"cidr": scanCidr}).Info("Randomizing permutations")
			permutations = utils.ShufflePermutations(permutations)
		}

		if err != nil {
			log.WithFields(log.Fields{
				"skip-http": skipHTTP, "skip-https": skipHTTPS, "ports": ports, "error": err,
			}).Fatal("Failed building permutations")
		}
		log.WithField("permutation-count", len(permutations)).Info("Total permutations to be processed")

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
			Total: len(permutations),
		}
		bar.SetStatus(status)
		bar.Render(os.Stdout)

		for _, permutation := range permutations {

			u, err := url.ParseRequestURI(permutation)
			if err != nil {

				log.WithField("url", permutation).Warn("Skipping Invalid URL")
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

		log.WithFields(log.Fields{"run-time": time.Since(startTime), "permutation-count": len(permutations)}).
			Info("Complete")
	},
}

// Validates that the arguments received for scanCmd is valid.
func validateScanCmdFlags() {

	// Ensure we have at least a CIDR
	if len(scanCidr) == 0 {
		log.WithField("cidr", scanCidr).Fatal("Please provide a CIDR scan")
	}

	if skipHTTP && skipHTTPS {
		log.WithFields(log.Fields{"skip-http": skipHTTP, "skip-https": skipHTTPS}).
			Fatal("Both HTTP and HTTPS cannot be skipped")

	}
}

func init() {
	RootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringSliceVarP(&scanCidr, "cidr", "c", []string{}, "The CIDR to scan (Can specify more than one --cidr)")
	scanCmd.Flags().BoolVarP(&skipHTTP, "no-http", "s", false, "Skip trying to connect with HTTP")
	scanCmd.Flags().BoolVarP(&skipHTTPS, "no-https", "S", false, "Skip trying to connect with HTTPS")
	scanCmd.Flags().StringVarP(&scanPorts, "ports", "p", "80,443,8080,8443", "Ports to scan")
	scanCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
	scanCmd.Flags().BoolVarP(&randomPermutations, "random", "r", false, "Randomize generated permutations")
}
