package utils

// SliceContainsInt checks if a slice has an int
func SliceContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
