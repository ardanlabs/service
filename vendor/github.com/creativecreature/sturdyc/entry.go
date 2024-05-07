package sturdyc

import "time"

// entry represents a single cache entry.
type entry[T any] struct {
	key                 string
	value               T
	expiresAt           time.Time
	refreshAt           time.Time
	numOfRefreshRetries int
	isMissingRecord     bool
}
