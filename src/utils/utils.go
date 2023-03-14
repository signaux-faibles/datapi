package utils

func Contains(array []string, test string) bool {
	for _, s := range array {
		if s == test {
			return true
		}
	}
	return false
}
