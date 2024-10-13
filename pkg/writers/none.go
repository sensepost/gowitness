package writers

import (
	"github.com/sensepost/gowitness/pkg/models"
)

// NoneWriter is a None writer
type NoneWriter struct {
}

// NewStdoutWriter initialises a none writer
func NewNoneWriter() (*NoneWriter, error) {
	return &NoneWriter{}, nil
}

// Write does nothing
func (s *NoneWriter) Write(result *models.Result) error {
	return nil
}
