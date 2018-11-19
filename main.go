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
func (t GlobalChaincode) history(token string) []byte {
	var filter = func(modification interface{}) bool {
		return true
	}
	var history = ParseHistory(t.GetHistoryForKey(token), filter)
	return ToJson(history)

}
func panicEmptyTokenParam(token string) {
	if token == "" {
		PanicString("param:token is empty")
	}
}
func panicEmptyTokenDataParam(tokenData string) []byte {
	if tokenData == "" {
		PanicString("param:tokenData is empty")
	}
	return []byte(tokenData)
}

//TODO use Token.hash as key
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
		panicEmptyTokenParam(tokenID)
		var tokenData TokenData //TODO nil check
		FromJson(panicEmptyTokenDataParam(params[1]), &tokenData)
		t.putToken(clientID, tokenID, tokenData)
	case Fcn_getToken:
		var tokenID = params[0]
		panicEmptyTokenParam(tokenID)
		responseBytes = t.getToken(clientID, tokenID)
	case Fcn_tokenHistory:
		var tokenID = params[0]
		panicEmptyTokenParam(tokenID)
		responseBytes = t.history(tokenID)
	case Fcn_deleteToken:
		var tokenID = params[0]
		panicEmptyTokenParam(tokenID)
		var tokenDataBytes = t.getToken(clientID, tokenID)

		if tokenDataBytes == nil {
			return //not exist, swallow
		}
		var tokenData TokenData
		FromJson(tokenDataBytes, &tokenData)
		if clientID.Cert.Subject.CommonName != tokenData.Owner {
			PanicString("Token Data Owner " + tokenData.Owner + " mismatched with CID.Subject.CN:" + clientID.Cert.Subject.CommonName)
		}
		t.DelState(tokenID)
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
