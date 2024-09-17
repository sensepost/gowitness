package islazy

// LeftTrucate a string if its more than max
func LeftTrucate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[max:]
}
