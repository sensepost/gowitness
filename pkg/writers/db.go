package writers

import (
	"sync"

	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm"
)

var hammingThreshold = 10

// DbWriter is a Database writer
type DbWriter struct {
	URI           string
	conn          *gorm.DB
	mutex         sync.Mutex
	hammingGroups []islazy.HammingGroup
}

// NewDbWriter initialises a database writer
func NewDbWriter(uri string, debug bool) (*DbWriter, error) {
	c, err := database.Connection(uri, false, debug)
	if err != nil {
		return nil, err
	}

	return &DbWriter{
		URI:           uri,
		conn:          c,
		mutex:         sync.Mutex{},
		hammingGroups: []islazy.HammingGroup{},
	}, nil
}

// Write results to the database
func (dw *DbWriter) Write(result *models.Result) error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	// Assign Group ID based on PerceptionHash
	groupID, err := dw.AssignGroupID(result.PerceptionHash)
	if err == nil {
		result.PerceptionHashGroupId = groupID
	} else {
		// if we couldn't get a perception hash, thats okay. maybe the
		// screenshot failed.
		log.Debug("could not get group id for perception hash", "hash", result.PerceptionHash)
	}

	return dw.conn.Create(result).Error
}

// AssignGroupID assigns a PerceptionHashGroupId based on Hamming distance
func (dw *DbWriter) AssignGroupID(perceptionHashStr string) (uint, error) {
	// Parse the incoming perception hash
	parsedHash, err := islazy.ParsePerceptionHash(perceptionHashStr)
	if err != nil {
		return 0, err
	}

	// Iterate through existing groups to find a match
	for _, group := range dw.hammingGroups {
		dist, err := islazy.HammingDistance(parsedHash, group.Hash)
		if err != nil {
			return 0, err
		}

		if dist <= hammingThreshold {
			return group.GroupID, nil
		}
	}

	// No matching group found; create a new group
	var maxGroupID uint
	err = dw.conn.Model(&models.Result{}).
		Select("COALESCE(MAX(perception_hash_group_id), 0)").
		Scan(&maxGroupID).Error
	if err != nil {
		return 0, err
	}
	nextGroupID := maxGroupID + 1

	// Add the new group to in-memory cache
	newGroup := islazy.HammingGroup{
		GroupID: nextGroupID,
		Hash:    parsedHash,
	}
	dw.hammingGroups = append(dw.hammingGroups, newGroup)

	return nextGroupID, nil
}
