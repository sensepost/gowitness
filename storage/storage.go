package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/tidwall/buntdb"
)

type Storage struct {
	Db *buntdb.DB
}

func (storage *Storage) Open(path string) error {

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

func (storage *Storage) SetHTTPData(data *HTTResponse) {

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

func (storage *Storage) Close() {

	log.Debug("Closing buntdb")
	storage.Db.Close()
}
