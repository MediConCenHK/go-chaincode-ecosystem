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

func (t GlobalChaincode) Put(cid ClientIdentity, txType string, params []string) {
	transient := t.GetTransient()
	switch txType {
	case "token":
		var payerAuth PayerAuth
		payerAuth = func(transient map[string][]byte) bool {
			return true
		}
		payerAuth.Exec(transient)
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
func (t GlobalChaincode) Transfer(cid ClientIdentity, txType string, params []string) []byte {
	switch txType {
	case "token":
		var tokenID = params[0]
		var from = params[1]
		var to = params[2]
		var toType = params[3]
		var tokenData TokenData
		var exist = t.GetStateObj(tokenID, &tokenData)
		if ! exist {
			PanicString("token " + tokenID + " not exist")
		}
		if tokenData.Owner != from {
			PanicString("token " + tokenID + " does not belong to " + from)
		}
		tokenData.Owner = to
		tokenData.OwnerType = tokenData.OwnerType.New(toType)
		t.PutStateObj(tokenID, tokenData)
	}
	return nil;
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
	case "transfer":
		t.Transfer(clientID, txType, params)
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
