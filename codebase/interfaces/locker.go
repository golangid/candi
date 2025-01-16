package interfaces

import (
	"time"

	"github.com/golangid/candi/options"
)

type (
	// Locker abstraction, lock concurrent process
	Locker interface {
		IsLocked(key string) bool
		IsLockedTTL(key string, ttl time.Duration) bool
		IsLockedWithOpts(key string, opts ...options.LockerOption) bool
		HasBeenLocked(key string) bool
		Unlock(key string)
		Reset(key string)
		Lock(key string, timeout time.Duration) (unlockFunc func(), err error)
		GetPrefixLocker() string
		GetTTLLocker() time.Duration
		Closer
	}
)
