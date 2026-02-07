package utils

import (
	"os"
)

func GetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func GetTextPreview(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
