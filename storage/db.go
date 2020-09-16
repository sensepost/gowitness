package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Db is the SQLite3 db handler ype
type Db struct {
	Path     string
	Disabled bool
}

// NewDb sets up a new DB
func NewDb() *Db {
	return &Db{}
}

// Get gets a db handle
func (db *Db) Get() (*gorm.DB, error) {

	if db.Disabled {
		return nil, nil
	}

	conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	conn.Logger.LogMode(logger.Silent)

	conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{})
	return conn, nil
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
