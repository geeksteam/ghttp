package moduleutils

import "strings"

// GetCurrentModule gets current panel module
func GetCurrentModule(path string) string {
	// Remove /api prefix
	path = strings.Split(path, "/")[2]
	// Remove GET paramaters from URI
	fields := strings.Split(path, "?")
	return fields[0]
}
