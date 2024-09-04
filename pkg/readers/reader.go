package readers

// Reader defines a reader
type Reader interface {
	Read(chan<- string) error
}
