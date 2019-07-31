//ECDH , inspired by go-ethereum
package crypto

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	. "github.com/davidkhala/goutils"
	"hash"
	"io"
	"math/big"
	"strconv"
)

type ECIESParams struct {
	Hash      func() hash.Hash // hash function
	hashAlgo  crypto.Hash
	Cipher    func([]byte) (cipher.Block, error) // symmetric cipher
	BlockSize int                                // block size of symmetric cipher
	KeyLen    int                                // length of symmetric key
}

// symDecrypt carries out CTR decryption using the block cipher specified in
// the parameters
func (params ECIESParams) SymDecrypt(key, cipherText []byte) (m []byte) {
	c, err := params.Cipher(key)
	PanicError(err)

	ctr := cipher.NewCTR(c, cipherText[:params.BlockSize])

	m = make([]byte, len(cipherText)-params.BlockSize)
	ctr.XORKeyStream(m, cipherText[params.BlockSize:])
	return
}

// symEncrypt carries out CTR encryption using the block cipher specified in the
// parameters.
func (params ECIESParams) SymEncrypt(rand io.Reader, key, m []byte) (cipherText []byte) {
	c, err := params.Cipher(key)
	PanicError(err)

	iv := make([]byte, params.BlockSize)
	_, err = io.ReadFull(rand, iv)
	PanicError(err)
	ctr := cipher.NewCTR(c, iv)

	cipherText = make([]byte, len(m)+params.BlockSize)
	copy(cipherText, iv)
	ctr.XORKeyStream(cipherText[params.BlockSize:], m)
	return
}

var (
	ECIES_AES128_SHA256 = ECIESParams{
		Hash:      sha256.New,
		hashAlgo:  crypto.SHA256,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    16,
	}

	ECIES_AES256_SHA256 = ECIESParams{
		Hash:      sha256.New,
		hashAlgo:  crypto.SHA256,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}

	ECIES_AES256_SHA384 = ECIESParams{
		Hash:      sha512.New384,
		hashAlgo:  crypto.SHA384,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}

	ECIES_AES256_SHA512 = ECIESParams{
		Hash:      sha512.New,
		hashAlgo:  crypto.SHA512,
		Cipher:    aes.NewCipher,
		BlockSize: aes.BlockSize,
		KeyLen:    32,
	}
)

// ParamsFromCurve selects parameters optimal for the selected elliptic curve.
// Only the curves P256, P384, and P512 are supported.
func ParamsFromCurve(curve elliptic.Curve) (params ECIESParams) {
	switch curve {
	case elliptic.P256():
		params = ECIES_AES128_SHA256
	case elliptic.P384():
		params = ECIES_AES256_SHA384
	case elliptic.P521():
		params = ECIES_AES256_SHA512
	default:
		PanicString("ecies: unsupported ECIES parameters")
	}
	return
}

// ECDH key agreement method used to establish secret keys for encryption.
func (prv ECPriv) GenerateShared(pub ECPub, skLen, macLen int) (sk []byte) {
	if prv.PublicKey.Curve != pub.Curve {
		PanicString("ecies: invalid elliptic curve")
	}
	if skLen+macLen > (pub.Curve.Params().BitSize+7)/8 {
		PanicString("ecies: shared key params are too big")
	}
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, prv.D.Bytes())
	if x == nil {
		PanicString("ecies: shared key is point at infinity")
	}
	sk = make([]byte, skLen+macLen)
	skBytes := x.Bytes()
	copy(sk[len(sk)-len(skBytes):], skBytes)
	return sk
}

func incCounter(ctr []byte) {
	if ctr[3]++; ctr[3] != 0 {
		return
	}
	if ctr[2]++; ctr[2] != 0 {
		return
	}
	if ctr[1]++; ctr[1] != 0 {
		return
	}
	if ctr[0]++; ctr[0] != 0 {
		return
	}
}

// NIST SP 800-56 Concatenation Key Derivation Function (see section 5.8.1).
func concatKDF(hash hash.Hash, z []byte, kdLen int) (k []byte) {

	reps := ((kdLen + 7) * 8) / (hash.BlockSize() * 8)
	var big2To32 = new(big.Int).Exp(big.NewInt(2), big.NewInt(32), nil) //TODO what is this
	var big2To32M1 = new(big.Int).Sub(big2To32, big.NewInt(1))          //TODO what is this
	if big.NewInt(int64(reps)).Cmp(big2To32M1) > 0 {
		PanicString("ecies: can't supply requested key data")
	}

	counter := []byte{0, 0, 0, 1}
	k = make([]byte, 0)

	for i := 0; i <= reps; i++ {
		hash.Write(counter)
		hash.Write(z)
		k = append(k, hash.Sum(nil)...)
		hash.Reset()
		incCounter(counter)
	}

	k = k[:kdLen]
	return
}

func messageTag(hash func() hash.Hash, km, msg []byte) []byte {
	mac := hmac.New(hash, km)
	mac.Write(msg)
	tag := mac.Sum(nil)
	return tag
}
func (pub ECPub) Encrypt(rand io.Reader, m []byte) (ct []byte) {
	params := ParamsFromCurve(pub.Curve)

	var newPriv = ECPriv{}.New(pub.Curve)

	hash := params.Hash()
	z := newPriv.GenerateShared(pub, params.KeyLen, params.KeyLen)
	K := concatKDF(hash, z, params.KeyLen+params.KeyLen)
	Ke := K[:params.KeyLen]
	Km := K[params.KeyLen:]
	hash.Write(Km)
	Km = hash.Sum(nil)
	hash.Reset()

	cipherText := params.SymEncrypt(rand, Ke, m)
	if len(cipherText) <= params.BlockSize {
		PanicString("ecies: cipher text is longer than params.BlockSize: " + strconv.Itoa(len(cipherText)) + " > " + strconv.Itoa(params.BlockSize))
	}

	d := messageTag(params.Hash, Km, cipherText)

	Rb := elliptic.Marshal(pub.Curve, newPriv.PublicKey.X, newPriv.PublicKey.Y)
	ct = make([]byte, len(Rb)+len(cipherText)+len(d))
	copy(ct, Rb)
	copy(ct[len(Rb):], cipherText)
	copy(ct[len(Rb)+len(cipherText):], d)
	return
}

// Decrypt decrypts an ECIES ciphertext.
func (prv ECPriv) Decrypt(c []byte) []byte {
	if len(c) == 0 {
		PanicString("ecies: invalid message")
	}
	var ecPub = ECPub{&prv.PublicKey}
	params := ParamsFromCurve(ecPub.Curve)
	hash := params.Hash()

	var (
		rLen   int
		hLen   = hash.Size()
		mStart int
		mEnd   int
	)

	switch c[0] {
	case 2, 3, 4:
		rLen = (prv.PublicKey.Curve.Params().BitSize + 7) / 4
		if len(c) < (rLen + hLen + 1) {
			PanicString("ecies: invalid message")
		}
	default:
		PanicString("ecies: invalid public key")
	}

	mStart = rLen
	mEnd = len(c) - hLen

	R := &ecdsa.PublicKey{}
	R.Curve = ecPub.Curve
	R.X, R.Y = elliptic.Unmarshal(R.Curve, c[:rLen])
	if R.X == nil {
		PanicString("ecies: invalid public key: point not on curve")
	}

	z := prv.GenerateShared(ECPub{R}, params.KeyLen, params.KeyLen)

	K := concatKDF(hash, z, params.KeyLen+params.KeyLen)

	Ke := K[:params.KeyLen]
	Km := K[params.KeyLen:]
	hash.Write(Km)
	Km = hash.Sum(nil)
	hash.Reset()

	d := messageTag(params.Hash, Km, c[mStart:mEnd])
	if subtle.ConstantTimeCompare(c[mEnd:], d) != 1 {
		PanicString("ecies: invalid message")
	}

	return params.SymDecrypt(Ke, c[mStart:mEnd])
}
