package readers

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/sensepost/gowitness/internal/islazy"
)

// NessusReader is a Nessus file reader
type NessusReader struct {
	Options *NessusReaderOptions
}

// NessusReaderOptions are options for a nessus file reader
type NessusReaderOptions struct {
	Source    string
	NoHTTP    bool
	NoHTTPS   bool
	Hostnames bool
	// filters
	Services      []string
	PluginOutputs []string
	PluginNames   []string
	Ports         []int
}

// structures for XML parsing
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

// NewNessusReader returns a new Nessus file reader
func NewNessusReader(opts *NessusReaderOptions) *NessusReader {
	return &NessusReader{
		Options: opts,
	}
}

func (nr *NessusReader) Read(ch chan<- string) error {
	defer close(ch)

	nessus, err := os.Open(nr.Options.Source)
	if err != nil {
		return err
	}
	defer nessus.Close()

	decoder := xml.NewDecoder(nessus)
	var targets = make(map[string][]int)

	for {
		token, err := decoder.Token()
		if err != nil || token == nil {
			break // EOF or error
		}

		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local != "ReportHost" {
				break
			}

			var host reportHost
			decoder.DecodeElement(&host, &element)

			var fqdn, ip string
			for _, v := range host.Tags {
				if v.Key == "host-fqdn" {
					fqdn = v.Value
				}
				if v.Key == "host-ip" {
					ip = v.Value
				}
			}

			for _, item := range host.ReportItems {
				// for future parsing debugging <3
				// log.Debug("report item", "ip", ip, "fqdn", fqdn)
				// log.Debug("detail", "plugin", item.PluginName, "service", item.ServiceName)

				// skip port if the port does not match the provided ports to filter
				if len(nr.Options.Ports) > 0 && !islazy.SliceHasInt(nr.Options.Ports, item.Port) {
					continue
				}

				// check the plugin name contains a given string. Contains should work, though startsWith may be useful.
				// A valid plugin name must be given here, otherwise we'll be iterating across too many pointless plugins.
				if !islazy.SliceHasStr(nr.Options.PluginNames, item.PluginName) {
					continue
				}

				// check the service name. typically this will at least be "web server" and or whatever plugin output
				if islazy.SliceHasStr(nr.Options.Services, item.ServiceName) ||
					islazy.SliceHasStr(nr.Options.PluginOutputs, item.PluginOutput) {

					// Add the hostnames or IP to the merged targetsMap
					if nr.Options.Hostnames && fqdn != "" {
						targets[fqdn] = islazy.UniqueIntSlice(append(targets[fqdn], item.Port))
					}
					if ip != "" {
						targets[ip] = islazy.UniqueIntSlice(append(targets[ip], item.Port))
					}
				}
			}
		}
	}

	for host, ports := range targets {
		for _, target := range nr.urlsFor(host, ports) {
			ch <- target
		}
	}

	return nil
}

// urlsFor generates urls for a target and its port ranges
func (nr *NessusReader) urlsFor(target string, ports []int) []string {
	var urls []string

	for _, port := range ports {
		if !nr.Options.NoHTTP {
			urls = append(urls, fmt.Sprintf("http://%s:%d", target, port))
		}

		if !nr.Options.NoHTTPS {
			urls = append(urls, fmt.Sprintf("https://%s:%d", target, port))
		}
	}

	return urls
}
