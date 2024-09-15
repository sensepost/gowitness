package readers

import (
	"bufio"
	"fmt"
	"net/url"
	"os"

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

	parsedURL, err := url.Parse(candidate)
	if err != nil {
		// invalid url, just bail
		return urls
	}

	// if the candidate already has a protocol defined, and there are no ports
	// to target, just return.
	// http here covers both http and https
	if parsedURL.Scheme != "" && parsedURL.Host != "" {
		// simplest return. no scheme needed, no ports needed
		if len(ports) == 0 {
			urls = append(urls, candidate)
			return urls
		} else {
			for _, port := range ports {
				parsedURL.Host = fmt.Sprintf("%s:%d", parsedURL.Hostname(), port)
				urls = append(urls, parsedURL.String())
			}
		}
	}

	// add a protocol, but respect the option to not add
	// either of http and https
	if !fr.Options.NoHTTP {
		if len(ports) > 0 {
			for _, port := range ports {
				rawURL := fmt.Sprintf("http://%s", candidate)
				parsedURL, err := url.Parse(rawURL)
				if err != nil {
					continue
				}
				parsedURL.Host = fmt.Sprintf("%s:%d", parsedURL.Hostname(), port)
				urls = append(urls, parsedURL.String())
			}
		} else {
			// just add the basic http URL
			urls = append(urls, "http://"+candidate)
		}
	}

	if !fr.Options.NoHTTPS {
		if len(ports) > 0 {
			for _, port := range ports {
				rawURL := fmt.Sprintf("https://%s", candidate)
				parsedURL, err := url.Parse(rawURL)
				if err != nil {
					continue
				}
				parsedURL.Host = fmt.Sprintf("%s:%d", parsedURL.Hostname(), port)
				urls = append(urls, parsedURL.String())
			}
		} else {
			urls = append(urls, "https://"+candidate)
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
