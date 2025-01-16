package options

import "time"

type (
	// Options for RedisLocker
	LockerOptions struct {
		Prefix string
		TTL    time.Duration
		Limit  int
	}

	// Option function type for setting options
	LockerOption func(*LockerOptions)
)
