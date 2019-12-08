package cmd

import (
	"bufio"
	"net/url"
	"os"
	"strings"
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

At least one --cidr flag or the --cidr-file flag (or both) should be specified.
If the subnet is omitted, it will be assumed that this is a /32. Multiple --cidr
flags are accepted.

When specifying the --random/-r flag, the ip:port permutations that are
generated will go through a shuffling phase so that the resultant
requests that are made wont follow each other on the same host.
This may be useful in cases where too many ports specified by the
--ports flag might trigger port scan alerts.

The --append-uri and --append-uri-file flags will generate extra permutations
where requests to these URI's will also be made. Consider an example network
range that would scan 192.168.0.1 on port 80, using '--append-uri foo' would
result in a permutation such as http://192.168.0.1:80/ including the permutation
http://192.168.0.1:80/foo.
Warning: Using a file to source URI's from using --append-uri-file with a large
amount of content will dramatically increase the number of permutations that
would be generated and utimately probed. Use with caution.

For example:

$ gowitness scan --cidr 192.168.0.0/24
$ gowitness scan --cidr 192.168.0.0/24 --cidr 10.10.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.1/32 --no-https
$ gowitness --log-level debug scan --threads 20 --ports 80,443,8080 --no-http --cidr 192.168.0.0/30
$ gowitness scan --ports 80,443,8080 --cidr 192.168.0.0/30 --append-uri '/admin'
$ gowitness scan --ports 80,443,8080 --cidr 192.168.0.0/30 --append-uri-file ~/wordlists/adminpanels.txt
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
		cidrs := readCidrs()
		log.WithField("cidr-count", len(cidrs)).Debug("Using CIDR ranges")

		// loop and parse the --cidr flags we got
		for _, cidr := range cidrs {

			if !strings.Contains(cidr, "/") {
				log.WithFields(log.Fields{"cidr": cidr}).Warning("CIDR does not have a subnet, assuming /32")
				cidr = cidr + "/32"
			}

			// parse the cidr
			cidrIps, err := utils.Hosts(cidr)
			if err != nil {
				log.WithFields(log.Fields{"cidr": cidr, "error": err}).Fatal("Failed to parse CIDR")
				return
			}

			// append the ips from the current cidr
			log.WithFields(log.Fields{"cidr": cidr, "cidr-ips": len(cidrIps)}).Debug("Appending cidr")
			ips = append(ips, cidrIps...)
		}

		log.WithFields(log.Fields{"total-ips": len(ips)}).Debug("Finished parsing CIDR ranges")

		permutations, err := utils.Permutations(ips, ports, skipHTTP, skipHTTPS)

		// uri appends
		if appendURI != "" {
			log.WithFields(log.Fields{"append-uri": appendURI}).Info("Appending URI to permutations")
			for _, permutation := range permutations {
				permutations = append(permutations, permutation+appendURI)
			}
		} else if appendURIFile != "" {
			var newPermutations []string
			log.WithFields(log.Fields{"append-uri-file": appendURIFile}).Info("Appending URIs from file")
			o, err := os.Open(appendURIFile)
			if err != nil {
				log.WithFields(log.Fields{"append-uri-file": appendURIFile, "error": err}).
					Fatal("Unable to open file to read URIs from")
			}
			scanner := bufio.NewScanner(o)
			for scanner.Scan() {
				u := scanner.Text()
				if !strings.HasPrefix(u, `/`) {
					log.WithField("append-uri-from-file", u).Debug("Prefixing file candidate with /")
					u = `/` + u
				}

				log.Debug(u)

				for _, permutation := range permutations {
					log.Debug(permutation, permutation+u)
					newPermutations = append(newPermutations, permutation+u)
				}
			}

			// figure out how to do this without yet another loop
			for _, k := range newPermutations {
				permutations = append(permutations, k)
			}
		}

		if randomPermutations {
			log.WithFields(log.Fields{"cidr-count": len(cidrs)}).Info("Randomizing permutations")
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

// populate the cidrs we are expecting from both the --cidr
// flags as well as when attempting to read a file from
// --file-cidr
func readCidrs() []string {

	var cidrs []string

	// read all of the --cidr flags
	for _, cidr := range scanCidr {
		cidrs = append(cidrs, cidr)
	}

	// read a file if one was specified
	if scanFileCidr != "" {

		file, err := os.Open(scanFileCidr)
		if err != nil {
			log.WithFields(log.Fields{"file": scanFileCidr, "err": err}).Fatal("Error reading CIDR file")
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			cidrs = append(cidrs, strings.TrimSpace(scanner.Text()))
		}
	}

	return cidrs
}

// Validates that the arguments received for scanCmd is valid.
func validateScanCmdFlags() {

	// Ensure we have at least a CIDR
	if len(scanCidr) == 0 && scanFileCidr == "" {
		log.WithFields(log.Fields{"cidr": scanCidr, "file-cidr": scanFileCidr}).
			Fatal("At least one --cidr or the --file-cidr flag is required")
	}

	// We need to have at least one protocol
	if skipHTTP && skipHTTPS {
		log.WithFields(log.Fields{"skip-http": skipHTTP, "skip-https": skipHTTPS}).
			Fatal("Both HTTP and HTTPS cannot be skipped")
	}

	if appendURI != "" && appendURIFile != "" {
		log.WithFields(log.Fields{"append-uri": appendURI, "append-uri-file": appendURIFile}).
			Fatal("Both --append-uri and --append-uri-file cannot be set")
	}

	if appendURI != "" {
		if !strings.HasPrefix(appendURI, `/`) {
			log.WithFields(log.Fields{"append-uri": `/`}).
				Warn("Append URI value does not start with a /, automatically appending it")
			appendURI = `/` + appendURI
		}
	}
}

func init() {
	RootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringSliceVarP(&scanCidr, "cidr", "c", []string{}, "The CIDR to scan (Can specify more than one --cidr)")
	scanCmd.Flags().StringVarP(&scanFileCidr, "file-cidr", "f", "", "A file containing newline separated CIDRs to scan")
	scanCmd.Flags().BoolVarP(&skipHTTP, "no-http", "s", false, "Skip trying to connect with HTTP")
	scanCmd.Flags().BoolVarP(&skipHTTPS, "no-https", "S", false, "Skip trying to connect with HTTPS")
	scanCmd.Flags().StringVarP(&scanPorts, "ports", "p", "80,443,8080,8443", "Ports to scan")
	scanCmd.Flags().IntVarP(&maxThreads, "threads", "t", 4, "Maximum concurrent threads to run")
	scanCmd.Flags().BoolVarP(&randomPermutations, "random", "r", false, "Randomize generated permutations")
	scanCmd.Flags().StringVarP(&appendURI, "append-uri", "a", "", "Add permutations appending this URI")
	scanCmd.Flags().StringVarP(&appendURIFile, "append-uri-file", "A", "", "Add permutations appending URI's from this file")
}
