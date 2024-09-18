package writers

import (
	"errors"
	"sync"

	"github.com/sensepost/gowitness/pkg/models"
)

// MemoryWriter is a memory-based results queue with a maximum slot count
type MemoryWriter struct {
	slots   int
	results []*models.Result
	mutex   sync.Mutex
}

// NewMemoryWriter initializes a MemoryWriter with the specified number of slots
func NewMemoryWriter(slots int) (*MemoryWriter, error) {
	if slots <= 0 {
		return nil, errors.New("slots need to be a positive integer")
	}

	return &MemoryWriter{
		slots:   slots,
		results: make([]*models.Result, 0, slots),
		mutex:   sync.Mutex{},
	}, nil
}

// Write adds a new result to the MemoryWriter.
func (s *MemoryWriter) Write(result *models.Result) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.results) >= s.slots {
		s.results = s.results[1:]
	}

	s.results = append(s.results, result)

	return nil
}

// GetLatest retrieves the most recently added result.
func (s *MemoryWriter) GetLatest() *models.Result {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.results) == 0 {
		return nil
	}

	return s.results[len(s.results)-1]
}

// GetFirst retrieves the oldest result in the MemoryWriter.
func (s *MemoryWriter) GetFirst() *models.Result {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.results) == 0 {
		return nil
	}

	return s.results[0]
}

// GetAllResults returns a copy of all current results.
func (s *MemoryWriter) GetAllResults() []*models.Result {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a copy to prevent external modification
	resultsCopy := make([]*models.Result, len(s.results))
	copy(resultsCopy, s.results)

	return resultsCopy
}
