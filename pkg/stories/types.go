package stories

// Story represents a user story.
type Story struct {
	ID              int `json:"id"`
	ExpiresAt       int `json:"expires_at"`
	DeviceTimestamp int `json:"device_timestamp"`
}
