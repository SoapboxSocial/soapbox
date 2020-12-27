package internal

// TrimRoomNameToLimit ensures the room name does not exceed 30 characters.
func TrimRoomNameToLimit(name string) string {
	if len([]rune(name)) > 30 {
		return string([]rune(name)[:30])
	}

	return name
}
