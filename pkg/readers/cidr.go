package readers

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
)

type CidrReader struct {
	Options *CidrReaderOptions
}

type CidrReaderOptions struct {
	NoHTTP      bool
	NoHTTPS     bool
	Cidrs       []string
	Source      string
	Ports       []int
	PortsSmall  bool
	PortsMedium bool
	PortsLarge  bool
	Random      bool
}

var (
	small  = []int{8080, 8443}
	medium = append(small, []int{81, 90, 591, 3000, 3128, 8000, 8008, 8081, 8082, 8834, 8888, 7015, 8800, 8990, 10000}...)
	large  = append(medium, []int{300, 2082, 2087, 2095, 4243, 4993, 5000, 7000, 7171, 7396, 7474, 8090, 8280, 8880, 9443}...)
)

func NewCidrReader(opts *CidrReaderOptions) *CidrReader {
	return &CidrReader{
		Options: opts,
	}
}

func (cr *CidrReader) Read(ch chan<- string) error {
	defer close(ch)

	candidates, err := cr.candidates()
	if err != nil {
		return err
	}

	log.Debug("total candidates to scan", "total", len(candidates))

	for _, target := range candidates {
		ch <- target
	}

	return nil
}

// candidates creates url candidates from ports and ips
func (cr *CidrReader) candidates() ([]string, error) {
	var candidates []string

	ports := cr.ports()
	ips, err := cr.ips()
	if err != nil {
		return nil, err
	}

	for _, ip := range ips {
		for _, port := range ports {
			partial := fmt.Sprintf("%s:%d", ip, port)

			if !cr.Options.NoHTTP {
				candidates = append(candidates, fmt.Sprintf("http://%s", partial))
			}

			if !cr.Options.NoHTTPS {
				candidates = append(candidates, fmt.Sprintf("https://%s", partial))
			}
		}
	}

	if cr.Options.Random {
		islazy.ShuffleStr(candidates)
	}

	return candidates, nil
}

// ports returns all of the ports to scan
func (cr *CidrReader) ports() []int {
	var ports = cr.Options.Ports

	if cr.Options.PortsSmall {
		ports = append(ports, small...)
	}

	if cr.Options.PortsMedium {
		ports = append(ports, medium...)
	}

	if cr.Options.PortsLarge {
		ports = append(ports, large...)
	}

	return islazy.UniqueIntSlice(ports)
}

// ips gets ips from a file and cidr agruments
func (cr *CidrReader) ips() ([]string, error) {
	var cidrs = cr.Options.Cidrs
	var ips []string

	// Slurp a file if we have one
	if cr.Options.Source != "" {
		var file *os.File
		var err error
		if cr.Options.Source == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(cr.Options.Source)
			if err != nil {
				return nil, err
			}
			defer file.Close()
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			cidrs = append(cidrs, strings.TrimSpace(scanner.Text()))
		}
	}

	// populate ips from the collected cidrs to return
	for _, cidr := range cidrs {
		if !strings.Contains(cidr, "/") {
			cidr += "/32"
		}

		ip, err := islazy.IpsInCIDR(cidr)
		if err != nil {
			return nil, err
		}

		ips = append(ips, ip...)
	}

	return ips, nil
}
