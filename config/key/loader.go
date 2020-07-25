package key

import (
	"crypto/rsa"
	"io/ioutil"

	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/dgrijalva/jwt-go"
)

type key struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func (k *key) PrivateKey() *rsa.PrivateKey {
	return k.private
}
func (k *key) PublicKey() *rsa.PublicKey {
	return k.public
}

// LoadRSAKey load rsa private key
func LoadRSAKey() interfaces.RSAKey {
	deferFunc := logger.LogWithDefer("Load RSA keys...")
	defer deferFunc()

	signBytes, err := ioutil.ReadFile("config/key/private.key")
	if err != nil {
		panic("Error when load private key. " + err.Error() + ". Please generate RSA keys")
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic("Error when load private key. " + err.Error())
	}

	verifyBytes, err := ioutil.ReadFile("config/key/public.pem")
	if err != nil {
		panic("Error when load public key. " + err.Error() + ". Please generate RSA keys")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic("Error when load public key. " + err.Error())
	}

	return &key{
		private: privateKey, public: publicKey,
	}
}
