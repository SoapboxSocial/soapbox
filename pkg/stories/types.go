package stories

// Story represents a user story.
type Story struct {
	ID              int   `json:"id"`
	ExpiresAt       int64 `json:"expires_at"`
	DeviceTimestamp int64 `json:"device_timestamp"`
}
