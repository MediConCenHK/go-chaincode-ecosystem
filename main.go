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
	t.Logger.Info("########### " + t.Name + " Init ###########")
	return shim.Success(nil)
}

func (t GlobalChaincode) putToken(cid ClientIdentity, params []string) {
	var tokenID = params[0]
	var tokenData TokenData
	FromJson([]byte(params[1]), &tokenData)
	tokenData.Client = cid
	t.PutStateObj(tokenID, tokenData)
}
func (t GlobalChaincode) getToken(cid ClientIdentity, params []string) []byte {
	var tokenID = params[0]
	var tokenData TokenData
	var exist = t.GetStateObj(tokenID, &tokenData)
	if ! exist {
		return nil
	}
	//TODO more logic here
	return ToJson(tokenData)
}
func (t GlobalChaincode) transferToken(cid ClientIdentity, params []string) []byte {
	var tokenID = params[0]
	var tokenTransferRequest TokenTransferRequest
	FromJson([]byte(params[1]), &tokenTransferRequest)
	var tokenData TokenData
	var exist = t.GetStateObj(tokenID, &tokenData)
	if ! exist {
		PanicString("token " + tokenID + " not exist")
	}
	if tokenData.Owner != tokenTransferRequest.FromOwner || tokenData.OwnerType != tokenTransferRequest.FromOwnerType {
		PanicString("token " + tokenID + " does not belong to [" + tokenTransferRequest.FromOwnerType.To() + "]" + tokenTransferRequest.FromOwner)
	}
	tokenData.Owner = tokenTransferRequest.ToOwner
	tokenData.OwnerType = tokenTransferRequest.ToOwnerType
	t.PutStateObj(tokenID, tokenData)
	return ToJson(tokenData)
}
func (t GlobalChaincode) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)

	var fcn, params = stub.GetFunctionAndParameters()
	t.Logger.Info("Invoke:fcn:" + fcn)
	var clientID = NewClientIdentity(stub)

	var transient = t.GetTransient()
	switch fcn {
	case Fcn_putToken:
		t.InsuranceAuth.Exec(transient)
		t.putToken(clientID, params)
		response = shim.Success(nil)
	case Fcn_getToken:
		var databytes = t.getToken(clientID, params)
		response = shim.Success(databytes)
	case Fcn_transferToken:
		t.InsuranceAuth.Exec(transient) //TODO modify case
		var databytes = t.transferToken(clientID, params)
		response = shim.Success(databytes)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	return
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
