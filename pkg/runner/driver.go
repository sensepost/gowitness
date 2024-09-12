package runner

// Driver is the interface browser drivers will implement.
type Driver interface {
	Witness(target string, runner *Runner)
	Close()
}
