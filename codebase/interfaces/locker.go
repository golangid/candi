package interfaces

import "time"

type (
	// Locker abstraction, lock concurrent process
	Locker interface {
		IsLocked(key string) bool
		HasBeenLocked(key string) bool
		Unlock(key string)
		Reset(key string)
		Lock(key string, timeout time.Duration) (unlockFunc func(), err error)
		Closer
	}
)
