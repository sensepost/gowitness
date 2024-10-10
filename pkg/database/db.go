package database

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connection returns a Database connection based on a URI
func Connection(uri string, shouldExist, debug bool) (*gorm.DB, error) {
	var err error
	var c *gorm.DB

	db, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var config = &gorm.Config{}
	if debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		config.Logger = logger.Default.LogMode(logger.Error)
	}

	switch db.Scheme {
	case "sqlite":
		if shouldExist {
			dbpath := filepath.Join(db.Host, db.Path)
			dbpath = filepath.Clean(dbpath)

			if _, err := os.Stat(dbpath); os.IsNotExist(err) {
				return nil, fmt.Errorf("sqlite database file does not exist: %s", dbpath)
			} else if err != nil {
				return nil, fmt.Errorf("error checking sqlite database file: %w", err)
			}
		}

		c, err = gorm.Open(sqlite.Open(db.Host+db.Path+"?cache=shared"), config)
		if err != nil {
			return nil, err
		}
		c.Exec("PRAGMA foreign_keys = ON")
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
		&models.Cookie{},
	); err != nil {
		return nil, err
	}

	return c, nil
}
