package storage

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Db is the SQLite3 db handler ype
type Db struct {
	Path          string
	SkipMigration bool

	// cli flags
	Disabled bool
	Debug    bool
	Platform int
}

const (
	Sqlite = iota
	Postgres
)

// NewDb sets up a new DB
func NewDb() *Db {
	return &Db{}
}

// Get gets a db handle
func (db *Db) Get() (*gorm.DB, error) {

	if db.Disabled {
		return nil, nil
	}

	var config = &gorm.Config{}
	if db.Debug {
		config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		config.Logger = logger.Default.LogMode(logger.Error)
	}

	switch db.Platform {
	case Sqlite:
		conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), config)
		if err != nil {
			return nil, err
		}

		if !db.SkipMigration {
			conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{}, &Technologie{}, &ConsoleLog{}, &NetworkLog{})
		}
		return conn, nil
	case Postgres:
		conn, err := gorm.Open(postgres.Open(db.Path), config)
		if err != nil {
			return nil, err
		}

		if !db.SkipMigration {
			conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{}, &Technologie{}, &ConsoleLog{}, &NetworkLog{})
		}
		return conn, nil
	}
	return nil, errors.New("invalid db platform")
}

// OrderPerception orders by perception hash if enabled
func OrderPerception(enabled bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if enabled {
			return db.Order("perception_hash desc")
		}
		return db
	}
}
