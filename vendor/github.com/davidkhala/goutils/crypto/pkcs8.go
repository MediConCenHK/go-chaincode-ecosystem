package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	. "github.com/davidkhala/goutils"
	"reflect"
)

//default in nodejs sdk
type PKCS8 struct {
	pem.Block
	Key interface{}
	reflect.Type
}

func (PKCS8) LoadPem(pemBytes []byte) (PKCS8) {
	block, rest := pem.Decode(pemBytes)
	assertEmpty(rest, "pem decode failed")
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	PanicError(err)
	return PKCS8{*block, privKey, reflect.TypeOf(privKey)}
}
func (t PKCS8) FormatECDSA() *ecdsa.PrivateKey {
	var result = t.Key.(*ecdsa.PrivateKey)
	return result
}

func (t PKCS8) ToPem() []byte {
	writer := bytes.NewBufferString("")
	err := pem.Encode(writer, &t.Block)
	PanicError(err)
	return writer.Bytes()
}
