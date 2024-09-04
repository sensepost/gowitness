package writers

import "github.com/sensepost/gowitness/pkg/models"

// Writer is a results writer
type Writer interface {
	Write(*models.Result) error
}
