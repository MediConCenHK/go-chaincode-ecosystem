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
	PayerAuth
	MemberAuth
}

func (t GlobalChaincode) Init(stub shim.ChaincodeStubInterface) (response peer.Response) {
	DeferPeerResponse(&response)
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
	DeferPeerResponse(&response)
	t.Prepare(stub)

	var fcn, params = stub.GetFunctionAndParameters()
	t.Logger.Info("Invoke:fcn:" + fcn)
	var clientID = NewClientIdentity(stub)

	var transient = t.GetTransient()
	switch fcn {
	case Fcn_putToken:
		t.PayerAuth.Exec(transient)
		t.putToken(clientID, params)
	case Fcn_getToken:
		if ! t.ClinicAuth(transient) && !t.MemberAuth(transient) && ! t.PayerAuth(transient) {
			PanicString("Identity authentication failed")
		}
		var databytes = t.getToken(clientID, params)
		return shim.Success(databytes)
	case Fcn_transferToken:
		t.PayerAuth.Exec(transient) //TODO modify case
		t.transferToken(clientID, params)
	default:
		PanicString("unknown fcn:" + fcn)
	}
	return shim.Success(nil)
}

func main() {
	var cc = GlobalChaincode{}
	cc.ClinicAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.PayerAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.MemberAuth = func(transient map[string][]byte) bool {
		return true
	}
	cc.SetLogger(GlobalCCID)
	shim.Start(cc)
}
