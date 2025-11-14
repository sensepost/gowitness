package database

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
		dsn, err := convertPostgresURItoDSN(uri)
		if err != nil {
			return nil, err
		}
		c, err = gorm.Open(postgres.Open(dsn), config)
		if err != nil {
			return nil, err
		}
	case "mysql":
		dsn, err := convertMySQLURItoDSN(uri)
		if err != nil {
			return nil, err
		}
		c, err = gorm.Open(mysql.Open(dsn), config)
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

func convertMySQLURItoDSN(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	user := parsed.User.Username()
	pass, _ := parsed.User.Password()
	host := parsed.Host
	dbname := strings.TrimPrefix(parsed.Path, "/")

	// Handle "tcp(...)"
	if strings.HasPrefix(host, "tcp(") && strings.HasSuffix(host, ")") {
		host = strings.TrimPrefix(host, "tcp(")
		host = strings.TrimSuffix(host, ")")
	}

	// Default port
	if !strings.Contains(host, ":") {
		host = host + ":3306"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, dbname,
	)

	return dsn, nil
}

func convertPostgresURItoDSN(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	user := parsed.User.Username()
	pass, _ := parsed.User.Password()
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "5432"
	}

	dbname := strings.TrimPrefix(parsed.Path, "/")

	// Start building the DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		host, user, pass, dbname, port,
	)

	// Add query params from URI
	query := parsed.Query()
	for key, values := range query {
		// Only take the first value per key
		dsn += fmt.Sprintf(" %s=%s", key, values[0])
	}

	return dsn, nil
}
