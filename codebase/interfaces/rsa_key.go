package interfaces

import "crypto/rsa"

// RSAKey abstraction
type RSAKey interface {
	PrivateKey() *rsa.PrivateKey
	PublicKey() *rsa.PublicKey
}
