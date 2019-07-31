package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	. "github.com/davidkhala/goutils"
	"math/big"
)

//Used for both ECDSA and ECDH
type ECPriv struct {
	*ecdsa.PrivateKey
}

//generate an EC private key (default to use P256 curve)
func (ECPriv) New(curve elliptic.Curve) (ECPriv) {
	if curve == nil {
		curve = elliptic.P256()
	}
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	PanicError(err)

	return ECPriv{priv}
}
func (t ECPriv) Sign(digest []byte) []byte {
	var r, s, err = ecdsa.Sign(rand.Reader, t.PrivateKey, digest)

	PanicError(err)
	return ECDSASignature{r, s}.Marshal()
}

type ECDSASignature struct {
	R, S *big.Int
}

func (t ECDSASignature) Marshal() []byte {
	var result, err = asn1.Marshal(t)
	PanicError(err)
	return result
}
func (ECDSASignature) Unmarshal(signature []byte) (ecdsaSignature ECDSASignature) {
	var rest, err = asn1.Unmarshal(signature, &ecdsaSignature)
	assertEmpty(rest, "asn1 unmarshal failed")
	PanicError(err)
	return
}
func (ECPriv) LoadPem(pemBytes []byte) (ECPriv) {
	block, rest := pem.Decode(pemBytes)
	assertEmpty(rest, "pem decode failed")
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	PanicError(err)
	return ECPriv{privKey}
}

func (t ECPriv) ToPem() []byte {
	writer := bytes.NewBufferString("")
	keyBytes, err := x509.MarshalECPrivateKey(t.PrivateKey)
	PanicError(err)
	pem.Encode(writer, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	return writer.Bytes()
}

// PublicKey is a representation of an elliptic curve public key.
type ECPub struct {
	*ecdsa.PublicKey
}

func (t ECPub) Verify(digest []byte, signature []byte) bool {
	var ecdsaSignature = ECDSASignature{}.Unmarshal(signature)
	return ecdsa.Verify(t.PublicKey, digest, ecdsaSignature.R, ecdsaSignature.S)
}
func (ECPub) LoadCert(pemBytes []byte) ECPub {
	block, rest := pem.Decode(pemBytes)
	assertEmpty(rest, "pem decode failed")
	var cert, err = x509.ParseCertificate(block.Bytes)
	PanicError(err)

	var pubkey = cert.PublicKey.(*ecdsa.PublicKey)
	return ECPub{pubkey}
}
