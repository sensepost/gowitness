package database

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	mysqldriver "github.com/go-sql-driver/mysql"
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
		// pgx parses URI-style DSNs natively, including percent-encoded
		// credentials and query parameters such as sslmode, so hand it the
		// URI unchanged rather than rebuilding it.
		c, err = gorm.Open(postgres.Open(uri), config)
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

	addr := parsed.Host

	// some users copy the tcp(...) form straight out of the driver's own dsn
	// documentation, so keep accepting it
	if strings.HasPrefix(addr, "tcp(") && strings.HasSuffix(addr, ")") {
		addr = strings.TrimSuffix(strings.TrimPrefix(addr, "tcp("), ")")
	}

	if !strings.Contains(addr, ":") {
		addr += ":3306"
	}

	password, _ := parsed.User.Password()

	cfg := mysqldriver.NewConfig()
	cfg.User = parsed.User.Username()
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = addr
	cfg.DBName = strings.TrimPrefix(parsed.Path, "/")
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Params = map[string]string{"charset": "utf8mb4"}

	// carry any parameters the user set on the uri, taking the first value
	// per key. these may override the defaults set above.
	for key, values := range parsed.Query() {
		cfg.Params[key] = values[0]
	}

	return cfg.FormatDSN(), nil
}
