package notifications

import "errors"

var (
	// ErrDeviceUnregistered is returned when an apns token is unregistered.
	ErrDeviceUnregistered = errors.New("apns device unregistered")

	// ErrRetryRequired is returned when a notification was not send due to a server and a retry is required.
	ErrRetryRequired = errors.New("failed to send retry required")
)
