package database

import (
	"errors"
	"net/url"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connection returns a Database connection based on a URI
func Connection(uri string, debug bool) (*gorm.DB, error) {
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
	case "mysql":
		c, err = gorm.Open(mysql.Open(uri), config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid db uri scheme")
	}

	// run database migrations on the connection
	if err := c.AutoMigrate(
		&models.Result{},
		&models.TLS{},
		&models.TLSSanList{},
		&models.Technology{},
		&models.Header{},
		&models.NetworkLog{},
		&models.ConsoleLog{},
	); err != nil {
		return nil, err
	}

	return c, nil
}
