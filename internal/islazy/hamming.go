package islazy

import (
	"encoding/hex"
	"errors"
	"strings"
)

// HammingGroup represents a hash -> group assignment used for
// inmemory hammingdistance calulations.
type HammingGroup struct {
	GroupID uint
	Hash    []byte
}

// HammingDistance calculates the number of differing bits between two byte slices.
func HammingDistance(hash1, hash2 []byte) (int, error) {
	if len(hash1) != len(hash2) {
		return 0, errors.New("hash lengths do not match")
	}

	distance := 0
	for i := 0; i < len(hash1); i++ {
		x := hash1[i] ^ hash2[i]
		for x != 0 {
			distance++
			x &= x - 1
		}
	}

	return distance, nil
}

// ParsePerceptionHash converts a perception hash string "p:<hex>" to a byte slice.
func ParsePerceptionHash(hashStr string) ([]byte, error) {
	if !strings.HasPrefix(hashStr, "p:") {
		return nil, errors.New("invalid perception hash format: missing 'p:' prefix")
	}

	hexPart := strings.TrimPrefix(hashStr, "p:")

	bytes, err := hex.DecodeString(hexPart)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
