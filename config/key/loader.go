package key

import (
	"crypto/rsa"
	"io/ioutil"

	"github.com/dgrijalva/jwt-go"
)

// LoadPrivateKey load rsa private key
func LoadPrivateKey() *rsa.PrivateKey {
	signBytes, err := ioutil.ReadFile("config/key/private.key")
	if err != nil {
		panic("Error when load private key. " + err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic("Error when load private key. " + err.Error())
	}
	return privateKey
}

// LoadPublicKey load rsa public key
func LoadPublicKey() *rsa.PublicKey {
	verifyBytes, err := ioutil.ReadFile("config/key/public.pem")
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}
	return publicKey
}
