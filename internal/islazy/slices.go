package islazy

// SliceHasStr checks if a slice has a string
func SliceHasStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

// SliceHasInt checks if a slice has an int
func SliceHasInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

// UniqueIntSlice returns a slice of unique ints
func UniqueIntSlice(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, num := range slice {
		if !seen[num] {
			seen[num] = true
			result = append(result, num)
		}
	}

	return result
}
