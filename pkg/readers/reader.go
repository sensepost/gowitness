package readers

// Reader defines a reader.
// NOTE: The Reader needs to close the channel when done to stop the runner.
// You would typically do this with a "defer close(ch)" at the start of your
// Read() implementation.
type Reader interface {
	Read(chan<- string) error
}
