package cmd

import (
	"net/url"
	"time"

	"github.com/remeh/sizedwaitgroup" // <3
	chrm "github.com/sensepost/gowitness/chrome"
	"github.com/sensepost/gowitness/utils"
	log "github.com/sirupsen/logrus"
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
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.1/32 --no-https
$ gowitness --log-level debug scan --threads 20 --ports 80,443,8080 --no-http --cidr 192.168.0.0/30

`,
	Run: func(cmd *cobra.Command, args []string) {

		validateScanCmdArguments()

		ports, _ := utils.Ports(scanPorts)
		log.WithField("ports", ports).Debug("Using ports")

		if len(ports) <= 0 {
			log.WithField("ports", scanPorts).Fatal("Please specify at least one port to connect to")
			return
		}

		ips, err := utils.Hosts(scanCidr)
		log.WithField("cidr", scanCidr).Debug("Using CIDR")

		if err != nil {
			log.WithFields(log.Fields{"cidr": scanCidr, "error": err}).Fatal("Failed to parse CIDR")
			return
		}

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
		chrome := chrm.InitChrome()

		// Set the screenshot path the user set.
		if err := chrome.ScreenshotPath(screenshotDestination); err != nil {

			log.WithFields(log.Fields{"error": err}).Fatal("Failed to set destination path")
			return
		}

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

				utils.ProcessURL(url, &chrome, waitTimeout)
			}(u)
		}

		swg.Wait()
		log.WithFields(log.Fields{"run-time": time.Since(startTime)}).Info("Complete")

	},
}

// Validates that the arguments received for scanCmd is valid.
func validateScanCmdArguments() {

	// Ensure we have at least a CIDR
	if scanCidr == "" {
		log.WithField("cidr", scanCidr).Fatal("Please provide a CIDR scan")
	}

	if skipHTTP && skipHTTPS {
		log.WithFields(log.Fields{"skip-http": skipHTTP, "skip-https": skipHTTPS}).
			Fatal("Both HTTP and HTTPS cannot be skipped")

	}
}

func init() {
	RootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanCidr, "cidr", "c", "", "The CIDR to scan")
	scanCmd.Flags().StringVarP(&screenshotDestination, "destination", "d", ".", "Destination directory for screenshots")
	scanCmd.Flags().BoolVarP(&skipHTTP, "no-http", "s", false, "Skip trying to connect with HTTP")
	scanCmd.Flags().BoolVarP(&skipHTTPS, "no-https", "S", false, "Skip trying to connect with HTTPS")
	scanCmd.Flags().StringVarP(&scanPorts, "ports", "p", "80,443,8008,8080", "Ports to scan")
	scanCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
	scanCmd.Flags().BoolVarP(&randomPermutations, "random", "r", false, "Randomize generated permutations")
}
