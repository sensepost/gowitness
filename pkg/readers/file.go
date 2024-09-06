package readers

import (
	"bufio"
	"os"
	"strings"
)

// FileReader is a reader that expects a file with targets that
// is newline delimited.
type FileReader struct {
	Options *FileReaderOptions
}

// FileReaderOptions are options for the file reader
type FileReaderOptions struct {
	Source  string
	NoHTTP  bool
	NoHTTPS bool
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		candidate := scanner.Text()
		if candidate == "" {
			continue
		}

		for _, url := range fr.urlsFor(candidate) {
			ch <- url
		}
	}

	return scanner.Err()
}

// urlsFor returns URLs for a scanning candidate.
// For candidates with no protocol, (and none of http/https is ignore), the
// method will return two urls
func (fr *FileReader) urlsFor(candidate string) []string {
	var urls []string
	// if the candidate already has a protocol defined, just return.
	// http here covers both http and https
	if strings.HasPrefix(candidate, "http") {
		urls = append(urls, candidate)
		return urls
	}

	// add a protocol, but respect the option to not add
	// either of http and https
	if !fr.Options.NoHTTP {
		urls = append(urls, "http://"+candidate)
	}

	if !fr.Options.NoHTTPS {
		urls = append(urls, "https://"+candidate)
	}

	return urls
}
