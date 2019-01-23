package main

import (
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/statebased"
)

func (t GlobalChaincode) addEPOrgsForKey(key string, orgs []string) {
	if len(orgs) < 1 {
		t.Logger.Warning("No EPOrgs to add for key [" + key + "]")
		return
	}

	t.Logger.Info("start add orgs for EP of key [", key, "], orgs: ", orgs)

	// get the endorsement policy for the key
	var epBytes []byte
	var err error
	epBytes, err = t.CCAPI.GetStateValidationParameter(key)

	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	ep, err := statebased.NewStateEP(epBytes)
	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	// add organizations to key level endorsement policy
	err = ep.AddOrgs(statebased.RoleTypePeer, orgs...)
	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	epBytes, err = ep.Policy()
	if err != nil {
		PanicString(err.Error())
	}

	// set the modified endorsement policy for the key
	err = t.CCAPI.SetStateValidationParameter(key, epBytes)
	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	t.Logger.Info("Successfully add orgs for EP of key [", key, "], orgs: ", orgs)

}

func (t GlobalChaincode) listOrgsForKeyEP(key string) []string{
	// get the endorsement policy for the key
	var epBytes []byte
	var err error
	epBytes, err = t.CCAPI.GetStateValidationParameter(key)

	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	ep, err := statebased.NewStateEP(epBytes)
	if err != nil {
		PanicString("failed to set key [" + key + "], err: " + err.Error())
	}

	// add organizations to key level endorsement policy
	orgs := ep.ListOrgs()
	t.Logger.Info("EP orgs: ", orgs, " for key[", key, "]")
	return orgs
}