package crypto

import (
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
)

// Ripemd160Hash 双哈希加密
func Ripemd160Hash(pubKey []byte) []byte {
	hash256 := sha256.New() // sha256哈希
	hash256.Write(pubKey)
	hash := hash256.Sum(nil)
	rmd160 := ripemd160.New() // ripemd哈希
	rmd160.Write(hash)
	return rmd160.Sum(nil)
}
