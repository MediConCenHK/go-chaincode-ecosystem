package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
)

func HashSha256(data []byte) []byte {
	var result = sha256.Sum256(data)
	return result[:]
}
func HashSha512(data []byte) []byte {
	var result = sha512.Sum512(data)
	return result[:]
}
