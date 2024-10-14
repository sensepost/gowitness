package readers

import (
	"fmt"
	"os"
	"strings"

	"github.com/lair-framework/go-nmap"
	"github.com/sensepost/gowitness/internal/islazy"
)

// NmapReader is an Nmap results reader
type NmapReader struct {
	Options *NmapReaderOptions
}

// NmapReaderOptions are options for the nmap reader
type NmapReaderOptions struct {
	// Path to an Nmap XML file
	Source  string
	NoHTTP  bool
	NoHTTPS bool
	// OpenOnly will only scan ports marked as open
	OpenOnly bool
	// Ports to limit scans to
	Ports []int
	// Ports to exclude, no matter what
	ExcludePorts []int
	// SkipPorts are ports to not scan
	SkipPorts []int
	// ServiceContains is a partial service filter
	ServiceContains string
	// Services is a service limit
	Services []string
	// Hostname is a hostname to use for url targets
	Hostnames bool
}

// NewNmapReader prepares a new Nmap reader
func NewNmapReader(opts *NmapReaderOptions) *NmapReader {
	return &NmapReader{
		Options: opts,
	}
}

// Read an nmap file
func (nr *NmapReader) Read(ch chan<- string) error {
	defer close(ch)

	xml, err := os.ReadFile(nr.Options.Source)
	if err != nil {
		return err
	}

	nmapXML, err := nmap.Parse(xml)
	if err != nil {
		return err
	}

	for _, host := range nmapXML.Hosts {
		for _, address := range host.Addresses {
			if !islazy.SliceHasStr([]string{"ipv4", "ipv6"}, address.AddrType) {
				continue
			}

			for _, port := range host.Ports {
				// filter only open ports
				if nr.Options.OpenOnly && port.State.State != "open" {
					continue
				}

				// if this port should always be excluded
				if len(nr.Options.ExcludePorts) > 0 && !islazy.SliceHasInt(nr.Options.ExcludePorts, port.PortId) {
					continue
				}

				// apply the port filter if it exists
				if len(nr.Options.Ports) > 0 && !islazy.SliceHasInt(nr.Options.Ports, port.PortId) {
					continue
				}

				// apply port skips
				if len(nr.Options.SkipPorts) > 0 && islazy.SliceHasInt(nr.Options.SkipPorts, port.PortId) {
					continue
				}

				// apply service filter
				if len(nr.Options.Services) > 0 && !islazy.SliceHasStr(nr.Options.Services, port.Service.Name) {
					continue
				}

				// apply partial service filter
				if len(nr.Options.ServiceContains) > 0 && !strings.Contains(nr.Options.ServiceContains, port.Service.Name) {
					continue
				}

				// filters are complete, generate urls to push into the channel

				// add hostname candidates
				if nr.Options.Hostnames {
					for _, hostaName := range host.Hostnames {
						for _, target := range nr.urlsFor(hostaName.Name, port.PortId) {
							ch <- target
						}
					}
				}

				// ip:port candidates
				if address.AddrType == "ipv4" {
					for _, target := range nr.urlsFor(address.Addr, port.PortId) {
						ch <- target
					}
				} else {
					addr := fmt.Sprintf("[%s]", address.Addr)
					for _, target := range nr.urlsFor(addr, port.PortId) {
						ch <- target
					}
				}
			}
		}
	}

	return nil
}

// urlsFor returns URLs for a scanning candidate.
// For candidates with no protocol, (and none of http/https is ignored), the
// method will return two urls
func (nr *NmapReader) urlsFor(target string, port int) []string {
	var urls []string

	if !nr.Options.NoHTTP {
		urls = append(urls, fmt.Sprintf("http://%s:%d", target, port))
	}

	if !nr.Options.NoHTTPS {
		urls = append(urls, fmt.Sprintf("https://%s:%d", target, port))
	}

	return urls
}
