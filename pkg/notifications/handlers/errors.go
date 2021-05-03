package handlers

import "errors"

var (
	errRoomPrivate   = errors.New("room is private")
	errNoRoomMembers = errors.New("room is empty")
	errFailedToSort  = errors.New("failed to sort")
	errEmptyResponse = errors.New("empty response")
	ErrNoCreator     = errors.New("no creator")
)
