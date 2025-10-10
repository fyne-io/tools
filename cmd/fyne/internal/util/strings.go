package util

func Contains(xs []string, a string) bool {
	for _, x := range xs {
		if x == a {
			return true
		}
	}
	return false
}
