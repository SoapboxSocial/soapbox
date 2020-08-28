package roomsV2

import "errors"

var (
	errInvalidSDP           = errors.New("invalid sdp")
	errSdpParseFailed       = errors.New("sdp parse failed")
	errConnectionInitFailed = errors.New("peer connection init failed")
)
