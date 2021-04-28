package notifications

import "errors"

var (
	// ErrDeviceUnregistered is returned when an apns token is unregistered.
	ErrDeviceUnregistered = errors.New("apns device unregistered")
)
