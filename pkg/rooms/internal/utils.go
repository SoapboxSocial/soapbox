package internal

import (
	"strings"

	"github.com/segmentio/ksuid"
)

// TrimRoomNameToLimit ensures the room name does not exceed 30 characters.
func TrimRoomNameToLimit(input string) string {
	name := strings.TrimSpace(input)
	if len([]rune(name)) > 30 {
		return string([]rune(name)[:30])
	}

	return name
}

// GenerateRoomID generates a random alpha-numeric room ID.
func GenerateRoomID() string {
	return strings.ToLower(ksuid.New().String())
}
