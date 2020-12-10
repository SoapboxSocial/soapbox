package stories

import "strings"

// IDFromName trims the file extension from the file and returns only the ID.
func IDFromName(name string) string {
	return strings.TrimSuffix(name, ".aac")
}
