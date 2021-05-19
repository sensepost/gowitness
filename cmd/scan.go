package cmd

import (
	"bufio"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sensepost/gowitness/lib"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a CIDR range and take screenshots along the way",
	Long: `Scans a CIDR range and takes screenshots along the way.

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
--ports flag might trigger port scan alerts.`,
	Example: `$ gowitness scan --cidr 192.168.0.0/24
$ gowitness scan --cidr 192.168.0.0/24 --cidr 10.10.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.0/24
$ gowitness scan --threads 20 --ports 80,443,8080 --cidr 192.168.0.1/32 --no-https
$ gowitness --log-level debug scan --threads 20 --ports 80,443,8080 --no-http --cidr 192.168.0.0/30`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		// prepare targets
		ports, err := getScanPorts()
		if err != nil {
			log.Fatal().Err(err).Msg("could not determine ports to scan")
		}
		log.Debug().Int("port-count", len(ports)).Msg("number of ports to scan")

		ips, err := getScanCidrIps()
		if err != nil {
			log.Fatal().Err(err).Msg("could not determine ports to scan")
		}
		log.Debug().Int("ip-count", len(ips)).Msg("number of ports to scan")

		if len(ports) == 0 || len(ips) == 0 {
			log.Warn().Int("ports", len(ports)).Int("ips", len("ips")).Msg("empty ports/ips determined. check flags")
		}

		targets, err := getScanPermutations(&ips, &ports)
		if err != nil {
			log.Fatal().Err(err).Msg("could not determine ports to scan")
		}
		log.Debug().Int("permutation-count", len(targets)).Msg("number of ports to scan")

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
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringSliceVarP(&options.ScanCidr, "cidr", "c", []string{}, "a cidr to scan (supports multiple --cidr flags)")
	scanCmd.Flags().StringVarP(&options.ScanCidrFile, "file-cidr", "f", "", "a file containing newline separated cidrs")
	scanCmd.Flags().BoolVar(&options.NoHTTPS, "no-https", false, "do not try using https://")
	scanCmd.Flags().BoolVar(&options.NoHTTP, "no-http", false, "do not try using http://")
	scanCmd.Flags().StringVar(&options.ScanPorts, "ports", "", "comma separated list of extra ports to scan")
	scanCmd.Flags().BoolVar(&options.PortsSmall, "ports-small", true, "also use the small ports list (80,443,8080,8443)")
	scanCmd.Flags().BoolVar(&options.PortsMedium, "ports-medium", false, "also use the medium ports list (small + 81,90,591,3000,3128,8000,8008,8081,8082,8834,8888,7015,8800,8990,10000)")
	scanCmd.Flags().BoolVar(&options.PortsLarge, "ports-large", false, "also use the large ports list (medium + 300,2082,2087,2095,4243,4993,5000,7000,7171,7396,7474,8090,8280,8880,9443)")
	scanCmd.Flags().IntVarP(&options.Threads, "threads", "t", 4, "threads used to run")
	scanCmd.Flags().BoolVarP(&options.ScanRandom, "random", "r", false, "randomize scan targets")
}

// getScanPorts determines all of the ports to use
func getScanPorts() ([]int, error) {

	portString := options.ScanPorts
	if !strings.HasSuffix(portString, ",") {
		portString += ","
	}

	if options.PortsSmall {
		portString += lib.PortsSmall
	}
	if options.PortsMedium {
		portString += lib.PortsMedium
	}
	if options.PortsLarge {
		portString += lib.PortsLarge
	}

	p, err := lib.PortsFromString(portString)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// getScanCidrIps returns a slice of all of the ips
// in a scan
func getScanCidrIps() (ips []string, err error) {

	var cidrs []string
	cidrs = append(cidrs, options.ScanCidr...)

	if options.ScanCidrFile != "" {
		file, err := os.Open(options.ScanCidrFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			cidrs = append(cidrs, strings.TrimSpace(scanner.Text()))
		}
	}

	for _, cidr := range cidrs {
		if !strings.Contains(cidr, "/") {
			cidr += "/32"
		}

		i, err := lib.HostsInCIDR(cidr)
		if err != nil {
			return nil, err
		}

		ips = append(ips, i...)
	}

	return
}

// getScanPermutations will generate url permutations from a port and ip slice.
// if random permutation order is needed, this function will take care of that
// too.
// todo: add uri appending support like we had in v1
func getScanPermutations(ips *[]string, ports *[]int) (results []string, err error) {

	for _, ip := range *ips {
		for _, port := range *ports {

			partialURL := ip + ":" + strconv.Itoa(port)
			if !options.NoHTTP {

				httpURL := "http://" + partialURL
				u, err := url.Parse(httpURL)
				if err != nil {
					return nil, err
				}

				results = append(results, u.String())
			}

			if !options.NoHTTPS {

				httpsURL := "https://" + partialURL
				u, err := url.Parse(httpsURL)
				if err != nil {
					return nil, err
				}

				results = append(results, u.String())
			}
		}
	}

	if options.ScanRandom {
		rand.Seed(time.Now().UTC().UnixNano())

		N := len(results)
		for i := 0; i < N; i++ {
			r := i + rand.Intn(N-i)
			results[r], results[i] = results[i], results[r]
		}
	}

	return
}
