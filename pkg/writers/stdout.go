package writers

import (
	"fmt"
	"os"

	"github.com/sensepost/gowitness/pkg/models"
)

// StdoutWriter is a Stdout writer
type StdoutWriter struct {
}

// NewStdoutWriter initialises a stdout writer
func NewStdoutWriter() (*StdoutWriter, error) {

	return &StdoutWriter{}, nil
}

// Write results to stdout
func (s *StdoutWriter) Write(result *models.Result) error {
	fmt.Fprintln(os.Stdout, result.URL)
	return nil
}
