package internal

import (
	"github.com/segmentio/ksuid"
)

// TrimRoomNameToLimit ensures the room name does not exceed 30 characters.
func TrimRoomNameToLimit(name string) string {
	if len([]rune(name)) > 30 {
		return string([]rune(name)[:30])
	}

	return name
}

// GenerateRoomID generates a random alpha-numeric room ID.
func GenerateRoomID() string {
	return ksuid.New().String()
}
