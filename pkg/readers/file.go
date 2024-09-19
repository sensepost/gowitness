package readers

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sensepost/gowitness/internal/islazy"
)

// FileReader is a reader that expects a file with targets that
// is newline delimited.
type FileReader struct {
	Options *FileReaderOptions
}

// FileReaderOptions are options for the file reader
type FileReaderOptions struct {
	Source      string
	NoHTTP      bool
	NoHTTPS     bool
	Ports       []int
	PortsSmall  bool
	PortsMedium bool
	PortsLarge  bool
	Random      bool
}

// NewFileReader prepares a new file reader
func NewFileReader(opts *FileReaderOptions) *FileReader {
	return &FileReader{
		Options: opts,
	}
}

// Read from a file that contains targets.
// FilePath can be "-" indicating that we should read from stdin.
func (fr *FileReader) Read(ch chan<- string) error {
	defer close(ch)

	var file *os.File
	var err error

	if fr.Options.Source == "-" {
		file = os.Stdin
	} else {
		file, err = os.Open(fr.Options.Source)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	// determine any ports
	ports := fr.ports()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		candidate := scanner.Text()
		if candidate == "" {
			continue
		}

		for _, url := range fr.urlsFor(candidate, ports) {
			ch <- url
		}
	}

	return scanner.Err()
}

// urlsFor returns URLs for a scanning candidate.
//
// For candidates with no protocol, (and none of http/https is ignored), the
// method will return two urls.
// If any ports configuration exists, those will also be added as candidates.
func (fr *FileReader) urlsFor(candidate string, ports []int) []string {
	var urls []string

	// trim any spaces
	candidate = strings.TrimSpace(candidate)

	// check if we got a scheme, add
	hasScheme := strings.Contains(candidate, "://")
	if !hasScheme {
		candidate = "http://" + candidate
	}

	parsedURL, err := url.Parse(candidate)
	if err != nil {
		// invalid url, return empty slice
		return urls
	}

	hasPort := parsedURL.Port() != ""
	hostname := parsedURL.Hostname()

	// if hostname is not set we may have rubbish input. try and "fix" it
	if hostname == "" {
		// is it hostname/path?
		if idx := strings.Index(candidate, "/"); idx != -1 {
			parsedURL.Host = candidate[:idx]
			parsedURL.Path = candidate[idx:]
			hostname = parsedURL.Hostname()
		} else {
			// its just a hostname then?
			parsedURL.Host = candidate
			parsedURL.Path = ""
			hostname = candidate
		}

		// at this point if hostname is still "", then just skip it entirely
		if hostname == "" {
			return urls
		}
	}

	if hasScheme && hasPort {
		// return the candidate as is
		urls = append(urls, parsedURL.String())
		return urls
	}

	// determine schemes to apply
	var schemes []string
	if hasScheme {
		schemes = append(schemes, parsedURL.Scheme)
	} else {
		if !fr.Options.NoHTTP {
			schemes = append(schemes, "http")
		}
		if !fr.Options.NoHTTPS {
			schemes = append(schemes, "https")
		}
	}

	// determine ports to use
	var targetPorts []int
	if hasPort {
		port, err := strconv.Atoi(parsedURL.Port())
		if err == nil { // just ignore it
			targetPorts = append(targetPorts, port)
		}
	} else {
		// If no port is specified, use the provided ports
		targetPorts = ports
	}

	// generate the urls
	for _, scheme := range schemes {
		for _, port := range targetPorts {
			host := hostname

			if port != 0 {
				if isIPv6(hostname) {
					host = fmt.Sprintf("[%s]:%d", hostname, port)
				} else {
					host = fmt.Sprintf("%s:%d", hostname, port)
				}
			}

			fullURL := url.URL{
				Scheme: scheme,
				Host:   host,
				Path:   parsedURL.Path,
			}

			urls = append(urls, fullURL.String())
		}
	}

	return urls
}

// ports returns all of the ports to scan
func (fr *FileReader) ports() []int {
	var ports = fr.Options.Ports

	if fr.Options.PortsSmall {
		ports = append(ports, small...)
	}

	if fr.Options.PortsMedium {
		ports = append(ports, medium...)
	}

	if fr.Options.PortsLarge {
		ports = append(ports, large...)
	}

	return islazy.UniqueIntSlice(ports)
}

func isIPv6(hostname string) bool {
	return len(hostname) > 0 && hostname[0] == '[' && hostname[len(hostname)-1] == ']'
}
