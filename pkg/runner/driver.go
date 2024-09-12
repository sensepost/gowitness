package runner

import "github.com/sensepost/gowitness/pkg/models"

// Driver is the interface browser drivers will implement.
type Driver interface {
	Witness(target string, runner *Runner) (*models.Result, error)
	Close()
}
