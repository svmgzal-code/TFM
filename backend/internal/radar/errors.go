package radar

import "errors"

// Sentinel errors returned by the radar layer.
var (
	// ErrAlreadyConnected is returned when connecting while not disconnected.
	ErrAlreadyConnected = errors.New("radar: already connected")
	// ErrNotConnected is returned when an operation requires an active connection.
	ErrNotConnected = errors.New("radar: not connected")
	// ErrTimeout wraps command timeouts.
	ErrTimeout = errors.New("radar: command timed out")
)
