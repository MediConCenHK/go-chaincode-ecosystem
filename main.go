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

func (t *GlobalChaincode) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	DeferPeerResponse(&response)
	t.Prepare(stub)
	t.Logger.Info("########### " + t.Name + " Init ###########")
	return shim.Success(nil)
}

func ClientAuth(cid ClientIdentity) bool {
	return true
}
func (t GlobalChaincode) Put(cid ClientIdentity, params []string) {
	var txType = params[0]

	switch txType {
	case "token":
		var tokenID = params[1]
		var tokenData TokenData
		FromJson([]byte(params[2]), &tokenData)
		t.PutStateObj(tokenID, tokenData)
	}
}
func (t GlobalChaincode) Get(cid ClientIdentity, params []string) []byte {
	var txType = params[0]
	switch txType {
	case "token":
		var tokenID = params[1]
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
func (t *GlobalChaincode) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	DeferPeerResponse(&response)
	t.Prepare(stub)
	t.Logger.Info("########### " + t.Name + " Invoke ###########")

	var fcn, params = stub.GetFunctionAndParameters()
	var clientID = NewClientIdentity(stub)


	switch strings.ToLower(fcn) {
	case "put":
		t.Put(clientID, params)
	case "get":
		var databytes = t.Get(clientID, params)
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
