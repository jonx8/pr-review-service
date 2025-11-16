package utils

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ReplaceInSlice(slice []string, old string, new string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		if item == old {
			result[i] = new
		} else {
			result[i] = item
		}
	}
	return result
}
