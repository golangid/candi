package key

import (
	"crypto/rsa"
	"io/ioutil"
	"log"

	"github.com/dgrijalva/jwt-go"
)

// LoadRSAKey load rsa private key
func LoadRSAKey(isUse bool) (*rsa.PrivateKey, *rsa.PublicKey) {
	if !isUse {
		return nil, nil
	}

	signBytes, err := ioutil.ReadFile("config/key/private.key")
	if err != nil {
		panic("Error when load private key. " + err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic("Error when load private key. " + err.Error())
	}

	verifyBytes, err := ioutil.ReadFile("config/key/public.pem")
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}

	log.Println("Success load RSA Key")
	return privateKey, publicKey
}
