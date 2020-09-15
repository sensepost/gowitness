package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	conn, err := gorm.Open(sqlite.Open(db.Path+"?cache=shared"), &gorm.Config{Logger: nil})
	if err != nil {
		return nil, err
	}

	conn.AutoMigrate(&URL{}, &Header{}, &TLS{}, &TLSCertificate{}, &TLSCertificateDNSName{})
	return conn, nil
}
