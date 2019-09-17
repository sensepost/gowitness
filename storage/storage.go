package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/tidwall/buntdb"
)

// Storage handles the pointer to a buntdb instance
type Storage struct {
	Db      *buntdb.DB
	Enabled bool
}

// Open creates a new connection to a buntdb database
func (storage *Storage) Open(path string) error {

	if !storage.Enabled {
		log.Fatalf("Tried to open db when it was disabled via flag")
		return nil
	}

	log.WithField("database-location", path).Debug("Opening buntdb")

	db, err := buntdb.Open(path)
	if err != nil {
		return err
	}

	// build some indexes
	db.CreateIndex("url", "*", buntdb.IndexJSON("url"))

	storage.Db = db

	return nil
}

// SetHTTPData stores HTTP information about a URL
func (storage *Storage) SetHTTPData(data *HTTResponse) {

	// do nothing if storage was disabled
	if !storage.Enabled {
		return
	}

	// marshal the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.WithField("err", err).Fatal("Error marshalling the HTTP response data to JSON")
	}

	// generate a key to use
	key := sha1.New()
	key.Write([]byte(data.URL))
	keyBytes := key.Sum(nil)
	keyString := hex.EncodeToString(keyBytes)
	log.WithFields(log.Fields{"url": data.URL, "key": keyString}).Debug("Calculated key for storage")

	// add the document
	err = storage.Db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(keyString, string(jsonData), nil)

		return err
	})

	if err != nil {
		log.WithField("err", err).Fatal("Error saving HTTP response data")
	}
}

// Close closes the connection to a buntdb connection
func (storage *Storage) Close() {

	if !storage.Enabled {
		return
	}

	log.Debug("Closing buntdb")
	storage.Db.Close()
}
