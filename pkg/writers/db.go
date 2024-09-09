package writers

import (
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm"
)

// DbWriter is a Database writer
type DbWriter struct {
	URI  string
	conn *gorm.DB
}

// NewDbWriter initialises a database writer
func NewDbWriter(uri string, debug bool) (*DbWriter, error) {
	c, err := database.Connection(uri, debug)
	if err != nil {
		return nil, err
	}

	return &DbWriter{
		URI:  uri,
		conn: c,
	}, nil
}

// Write results to the database
func (dw *DbWriter) Write(result *models.Result) error {
	return dw.conn.Create(result).Error
}
