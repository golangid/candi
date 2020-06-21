package interfaces

import "crypto/rsa"

// Key abstraction
type Key interface {
	PrivateKey() *rsa.PrivateKey
	PublicKey() *rsa.PublicKey
}
