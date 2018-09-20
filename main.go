package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/fabric-common-chaincode-golang"
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"strings"
)

type GlobalChaincode struct {
	CommonChaincode
}

func (t GlobalChaincode) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	DeferPeerResponse(&response)
	t.Prepare(stub)
	t.Logger.Info("########### " + t.Name + " Init ###########")
	return shim.Success(nil)
}

func ClientAuth(cid ClientIdentity) bool {
	return true
}
func (t GlobalChaincode) Put(cid ClientIdentity, txType string, params []string) {
	switch txType {
	case "token":
		var tokenID = params[0]
		var tokenData TokenData
		FromJson([]byte(params[1]), &tokenData)
		t.PutStateObj(tokenID, tokenData)
	default:
		PanicString("unknown txType:" + txType)
	}
}
func (t GlobalChaincode) Get(cid ClientIdentity, txType string, params []string) []byte {
	switch txType {
	case "token":
		var tokenID = params[0]
		var tokenData TokenData
		var exist = t.GetStateObj(tokenID, &tokenData)
		if ! exist {
			return nil
		}
		//TODO more logic here
		return ToJson(tokenData)
	}
	return nil
}
func (t GlobalChaincode) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	DeferPeerResponse(&response)
	t.Prepare(stub)

	var fcn, params = stub.GetFunctionAndParameters()
	var txType = params[0]
	t.Logger.Info("Invoke:fcn:" + fcn + " txType:" + txType)
	params = params[1:]
	var clientID = NewClientIdentity(stub)

	switch strings.ToLower(fcn) {
	case "put":
		t.Put(clientID, txType, params)
	case "get":
		var databytes = t.Get(clientID, txType, params)
		return shim.Success(databytes)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	return shim.Success(nil)
}

func main() {
	var cc = GlobalChaincode{}
	cc.SetLogger("Global")
	shim.Start(cc)
}
