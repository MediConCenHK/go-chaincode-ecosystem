package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/fabric-common-chaincode-golang"
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type GlobalChaincode struct {
	CommonChaincode
	ClinicAuth
	NetworkAuth
	MemberAuth
	InsuranceAuth
}

func (t GlobalChaincode) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)
	t.Logger.Info("Init")
	return shim.Success(nil)
}

func (t GlobalChaincode) putToken(cid ClientIdentity, tokenID string, tokenData TokenData) {
	tokenData.Client = cid
	t.PutStateObj(tokenID, tokenData)
}
func (t GlobalChaincode) getToken(cid ClientIdentity, token string) []byte {
	var tokenData TokenData
	var exist = t.GetStateObj(token, &tokenData)
	if ! exist {
		return nil
	}
	return ToJson(tokenData)
}
func (t GlobalChaincode) transferToken(cid ClientIdentity, token string, request TokenTransferRequest) []byte {

	var tokenData TokenData
	var exist = t.GetStateObj(token, &tokenData)
	if ! exist {
		PanicString("token " + token + " not exist")
	}
	if tokenData.Owner != request.FromOwner || tokenData.OwnerType != request.FromOwnerType {
		PanicString("token " + token + " does not belong to [" + request.FromOwnerType.To() + "]" + request.FromOwner)
	}
	tokenData.Owner = request.ToOwner
	tokenData.OwnerType = request.ToOwnerType
	tokenData.Client = cid
	t.PutStateObj(token, tokenData)
	return ToJson(tokenData)
}
func (t GlobalChaincode) history(token string) []byte {
	var filter = func(modification interface{}) bool {
		return true
	}
	var history = ParseHistory(t.GetHistoryForKey(token), filter)
	return ToJson(history)

}
func panicToken(token string) {
	if token == "" {
		PanicString("token is empty")
	}
}
func panicTokenData(tokenData string) []byte {
	if tokenData == "" {
		PanicString("tokenData is empty")
	}
	return []byte(tokenData)
}
func (t GlobalChaincode) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)

	var fcn, params = stub.GetFunctionAndParameters()
	t.Logger.Info("Invoke:fcn", fcn)
	t.Logger.Debug("Invoke:params", params)
	var clientID = NewClientIdentity(stub)
	var transient = t.GetTransient()
	var responseBytes []byte
	const Fcn_tokenHistory = "tokenHistory"
	switch fcn {
	case Fcn_putToken:
		t.InsuranceAuth.Exec(transient)
		var tokenID = params[0] //TODO nil check
		panicToken(tokenID)
		var tokenData TokenData //TODO nil check
		FromJson(panicTokenData(params[1]), &tokenData)
		t.putToken(clientID, tokenID, tokenData)
	case Fcn_getToken:
		var tokenID = params[0]
		panicToken(tokenID)
		responseBytes = t.getToken(clientID, tokenID)
	case Fcn_transferToken:
		t.InsuranceAuth.Exec(transient) //TODO modify case
		var tokenID = params[0]
		panicToken(tokenID)
		var tokenTransferRequest TokenTransferRequest
		FromJson(panicTokenData(params[1]), &tokenTransferRequest)
		responseBytes = t.transferToken(clientID, tokenID, tokenTransferRequest)
	case Fcn_tokenHistory:
		var tokenID = params[0]
		panicToken(tokenID)
		responseBytes = t.history(tokenID)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	t.Logger.Debug("response", string(responseBytes))
	return shim.Success(responseBytes)
}

func main() {
	var cc = GlobalChaincode{}
	cc.ClinicAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.InsuranceAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.NetworkAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.MemberAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.SetLogger(GlobalCCID)
	shim.Start(cc)
}
