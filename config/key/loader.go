package key

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go"
)

// LoadPrivateKey load rsa private key
func LoadPrivateKey() *rsa.PrivateKey {
	signBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/config/key/private.key", os.Getenv("APP_PATH")))
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
	verifyBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/config/key/public.pem", os.Getenv("APP_PATH")))
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}
	return publicKey
}
