package readers

// Reader defines a reader.
// NOTE: The Reader needs to close the channel when done to stop the runner.
// You would typically do this with a "defer close(ch)" at the start of your
// Read() implementation.
type Reader interface {
	Read(chan<- string) error
}

// port collections that readers can refer to
var (
	small  = []int{8080, 8443}
	medium = append(small, []int{81, 90, 591, 3000, 3128, 8000, 8008, 8081, 8082, 8834, 8888, 7015, 8800, 8990, 10000}...)
	large  = append(medium, []int{300, 2082, 2087, 2095, 4243, 4993, 5000, 7000, 7171, 7396, 7474, 8090, 8280, 8880, 9443}...)
)
