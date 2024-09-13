package islazy

import "time"

// Float64ToTime takes a float64 as number of seconds since unix epoch and returns time.Time
//
// example field where this is used (expires field):
//
//	https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-Cookie
func Float64ToTime(f float64) time.Time {
	if f == 0 {
		// Return zero value for session cookies
		return time.Time{}
	}
	return time.Unix(0, int64(f*float64(time.Second)))
}
