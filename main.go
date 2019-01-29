package main

import (
	. "github.com/MediConCenHK/go-chaincode-common"
	. "github.com/davidkhala/fabric-common-chaincode-golang"
	. "github.com/davidkhala/fabric-common-chaincode-golang/cid"
	"github.com/davidkhala/fabric-common-chaincode-golang/ext"
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
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
func (t GlobalChaincode) getToken(token string) *TokenData {
	var tokenData TokenData
	var exist = t.GetStateObj(token, &tokenData)
	if ! exist {
		return nil
	}
	return &tokenData
}
func (t GlobalChaincode) history(token string) []byte {
	var history = ParseHistory(t.GetHistoryForKey(token), nil)
	return ToJson(history)

}

func (t GlobalChaincode) Invoke(stub shim.ChaincodeStubInterface) (response peer.Response) {
	defer Deferred(DeferHandlerPeerResponse, &response)
	t.Prepare(stub)

	var fcn, params = stub.GetFunctionAndParameters()
	t.Logger.Info("Invoke:fcn", fcn)
	t.Logger.Debug("Invoke:params", params)
	var clientID = NewClientIdentity(stub)
	var responseBytes []byte
	var tokenRaw = params[0]
	if tokenRaw == "" {
		PanicString("param:token is empty")
	}
	var tokenID = Hash([]byte(tokenRaw))

	var tokenData TokenData
	switch fcn {
	case Fcn_putToken:
		FromJson([]byte(params[1]), &tokenData)
		var tokenDataPtr = t.getToken(tokenID)
		//fixme identity verification
		if tokenDataPtr != nil {
			PanicString("token[" + tokenRaw + "] already exist")
		}
		tokenData.OwnerType = OwnerTypeMember
		tokenData.TransferDate = TimeLong(0)
		t.putToken(clientID, tokenID, tokenData)
		var keyPolicy = ext.NewKeyEndorsementPolicy(nil)

		keyPolicy.AddOrgs(msp.MSPRole_MEMBER, clientID.MspID)

		t.SetStateValidationParameter(tokenID, keyPolicy.Policy())

	case Fcn_getToken:
		var tokenDataPtr = t.getToken(tokenID)
		if tokenDataPtr == nil {
			break
		}
		responseBytes = ToJson(*tokenDataPtr)
	case Fcn_renewToken:
		var newExpiryTime = ParseTime(params[1])
		var tokenDataPtr = t.getToken(tokenID)
		if tokenDataPtr == nil {
			PanicString("token[" + tokenRaw + "] not found:")
		}
		tokenData = *tokenDataPtr
		tokenData.ExpiryDate = newExpiryTime
		t.putToken(clientID, tokenID, tokenData)
	case Fcn_tokenHistory:
		responseBytes = t.history(tokenID)
	case Fcn_deleteToken:
		var tokenDataPtr = t.getToken(tokenID)
		if tokenDataPtr == nil {
			break //not exist, swallow
		}
		tokenData = *tokenDataPtr
		if clientID.MspID != tokenData.Manager {
			PanicString("[" + tokenRaw + "]Token Data Manager(" + tokenData.Manager + ") mismatched with tx creator MspID: " + clientID.MspID)
		}
		t.DelState(tokenID)
	case Fcn_moveToken:
		var transferReq TokenTransferRequest

		FromJson([]byte(params[1]), &transferReq)

		var tokenDataPtr = t.getToken(tokenID)
		if tokenDataPtr == nil {
			PanicString("token[" + tokenRaw + "] not found:")
		}
		tokenData = *tokenDataPtr
		if tokenData.OwnerType != OwnerTypeMember {
			PanicString("original token OwnerType should be member, but got " + tokenData.OwnerType.To())
		}
		if tokenData.TransferDate != TimeLong(0) {
			PanicString("token already transferred")
		}

		tokenData = transferReq.ApplyOn(tokenData)
		tokenData.OwnerType = OwnerTypeNetwork
		tokenData.TransferDate = UnixMilliSecond(t.GetTxTime())
		tokenData.MetaData = transferReq.MetaData
		t.putToken(clientID, tokenID, tokenData)
		var keyPolicyBytes = t.GetStateValidationParameter(tokenID)
		var keyPolicy = ext.NewKeyEndorsementPolicy(keyPolicyBytes)
		keyPolicy.AddOrgs(msp.MSPRole_MEMBER, clientID.MspID)
		t.SetStateValidationParameter(tokenID, keyPolicy.Policy())
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
	ChaincodeStart(cc)
}
