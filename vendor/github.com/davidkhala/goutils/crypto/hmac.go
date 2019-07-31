package crypto

import (
	"crypto/hmac"
	"hash"
)

func CheckHMAC(h func() hash.Hash, message, messageMAC, key []byte) bool {
	mac := hmac.New(h, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
func NewHMAC(h func() hash.Hash, message, key []byte) []byte {
	mac := hmac.New(h, key)
	mac.Write(message)
	return mac.Sum(nil)
}
