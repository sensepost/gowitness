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
