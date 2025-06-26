package globals

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
)

var (
	RsaPrivateKey *rsa.PrivateKey
	RsaPublicKey  []byte
)

func init() {
	var err error
	RsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic("failed to generate RSA key")
	}
	RsaPublicKey = x509.MarshalPKCS1PublicKey(&RsaPrivateKey.PublicKey)
}
