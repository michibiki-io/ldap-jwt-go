package model

import (
	"crypto/rsa"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type Rsa struct {
	VerifyKey *rsa.PublicKey
	SignKey   *rsa.PrivateKey
}

func (rsa *Rsa) Load(privateKeyPath, publicKeyPath string) error {
	signBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	rsa.SignKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}

	verifyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}
	rsa.VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return err
	}

	return nil
}
