package main

import (
	"fmt"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/crypto"
	"testing"
)

func Test(t *testing.T) {
	var owner = "H008800.GP/0008H08"
	var ownerHash = crypto.HashSha256([]byte(owner)) //some chars are now allowed in Cert CommonName
	var hashHex = goutils.HexEncode(ownerHash)[32:]
	fmt.Print(hashHex)
}
