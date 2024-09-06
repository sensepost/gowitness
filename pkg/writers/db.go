package writers

import (
	"errors"
	"net/url"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DbWriter is a Database writer
type DbWriter struct {
	URI  string
	conn *gorm.DB
}

// NewDbWriter initialises a database writer
func NewDbWriter(uri string, debug bool) (*DbWriter, error) {
	var err error
	var c *gorm.DB

	db, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var config = &gorm.Config{}
	if debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	switch db.Scheme {
	case "sqlite":
		c, err = gorm.Open(sqlite.Open(db.Host+db.Path+"?cache=shared"), config)
		if err != nil {
			return nil, err
		}
	case "postgres":
		c, err = gorm.Open(postgres.Open(uri), config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid db uri scheme")
	}

	// run database migrations on the connection
	if err := c.AutoMigrate(
		&models.Result{},
		&models.Header{},
		&models.NetworkLog{},
		&models.ConsoleLog{},
	); err != nil {
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
