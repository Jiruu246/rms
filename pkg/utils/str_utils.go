package utils

import "strings"

func JoinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(slice[0])
	for i := 1; i < len(slice); i++ {
		result.WriteString(separator + slice[i])
	}
	return result.String()
}
