package scenjsonparse

import (
	"errors"
	"fmt"

	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	oj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/orderedjson"
)

func (p *Parser) processCheckDCDTData(
	tokenName mj.JSONBytesFromString,
	dcdtDataRaw oj.OJsonObject) (*mj.CheckDCDTData, error) {

	switch data := dcdtDataRaw.(type) {
	case *oj.OJsonString:
		// simple string representing balance "400,000,000,000"
		dcdtData := mj.CheckDCDTData{
			TokenIdentifier: tokenName,
		}
		balance, err := p.processCheckBigInt(dcdtDataRaw, bigIntUnsignedBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid DCDT balance: %w", err)
		}
		dcdtData.Instances = []*mj.CheckDCDTInstance{
			{
				Nonce:   mj.JSONCheckUint64{Value: 0, Original: ""},
				Balance: balance,
			},
		}
		return &dcdtData, nil
	case *oj.OJsonMap:
		return p.processCheckDCDTDataMap(tokenName, data)
	default:
		return nil, errors.New("invalid JSON object for DCDT")
	}
}

// map containing other fields too, e.g.:
//
//	{
//		"balance": "400,000,000,000",
//		"frozen": "true"
//	}
func (p *Parser) processCheckDCDTDataMap(tokenName mj.JSONBytesFromString, dcdtDataMap *oj.OJsonMap) (*mj.CheckDCDTData, error) {
	dcdtData := mj.CheckDCDTData{
		TokenIdentifier: tokenName,
	}
	// var err error
	firstInstance := &mj.CheckDCDTInstance{
		Nonce:      mj.JSONCheckUint64Unspecified(),
		Balance:    mj.JSONCheckBigIntUnspecified(),
		Creator:    mj.JSONCheckBytesUnspecified(),
		Royalties:  mj.JSONCheckUint64Unspecified(),
		Hash:       mj.JSONCheckBytesUnspecified(),
		Uri:        mj.JSONCheckBytesUnspecified(),
		Attributes: mj.JSONCheckBytesUnspecified(),
	}
	firstInstanceLoaded := false
	var explicitInstances []*mj.CheckDCDTInstance

	for _, kvp := range dcdtDataMap.OrderedKV {
		// it is allowed to load the instance directly, fields set to the first instance
		instanceFieldLoaded, err := p.tryProcessCheckDCDTInstanceField(kvp, firstInstance)
		if err != nil {
			return nil, fmt.Errorf("invalid account DCDT instance field: %w", err)
		}
		if instanceFieldLoaded {
			firstInstanceLoaded = true
		} else {
			switch kvp.Key {
			case "instances":
				explicitInstances, err = p.processCheckDCDTInstances(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT instances: %w", err)
				}
			case "lastNonce":
				dcdtData.LastNonce, err = p.processCheckUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT lastNonce: %w", err)
				}
			case "roles":
				dcdtData.Roles, err = p.processStringList(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT roles: %w", err)
				}
			case "frozen":
				dcdtData.Frozen, err = p.processCheckUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid DCDT frozen flag: %w", err)
				}
			default:
				return nil, fmt.Errorf("unknown DCDT data field: %s", kvp.Key)
			}
		}
	}

	if firstInstanceLoaded {
		dcdtData.Instances = []*mj.CheckDCDTInstance{firstInstance}
	}
	dcdtData.Instances = append(dcdtData.Instances, explicitInstances...)

	return &dcdtData, nil
}

func (p *Parser) tryProcessCheckDCDTInstanceField(kvp *oj.OJsonKeyValuePair, targetInstance *mj.CheckDCDTInstance) (bool, error) {
	var err error
	switch kvp.Key {
	case "nonce":
		targetInstance.Nonce, err = p.processCheckUint64(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid account nonce: %w", err)
		}
	case "balance":
		targetInstance.Balance, err = p.processCheckBigInt(kvp.Value, bigIntUnsignedBytes)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT balance: %w", err)
		}
	case "creator":
		targetInstance.Creator, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT creator address: %w", err)
		}
	case "royalties":
		targetInstance.Royalties, err = p.processCheckUint64(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT royalties: %w", err)
		}
		if targetInstance.Royalties.Value > 10000 {
			return false, errors.New("invalid DCDT NFT royalties: value exceeds maximum allowed 10000")
		}
	case "hash":
		targetInstance.Hash, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT hash: %w", err)
		}
	case "uri":
		targetInstance.Uri, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT URI: %w", err)
		}
	case "attributes":
		targetInstance.Attributes, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT attributes: %w", err)
		}
	default:
		return false, nil
	}
	return true, nil
}

func (p *Parser) processCheckDCDTInstances(dcdtInstancesRaw oj.OJsonObject) ([]*mj.CheckDCDTInstance, error) {
	var instancesResult []*mj.CheckDCDTInstance
	dcdtInstancesList, isList := dcdtInstancesRaw.(*oj.OJsonList)
	if !isList {
		return nil, errors.New("dcdt instances object is not a list")
	}
	for _, instanceItem := range dcdtInstancesList.AsList() {
		instanceAsMap, isMap := instanceItem.(*oj.OJsonMap)
		if !isMap {
			return nil, errors.New("JSON map expected as dcdt instances list item")
		}

		instance := mj.NewCheckDCDTInstance()

		for _, kvp := range instanceAsMap.OrderedKV {
			instanceFieldLoaded, err := p.tryProcessCheckDCDTInstanceField(kvp, instance)
			if err != nil {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: %w", err)
			}
			if !instanceFieldLoaded {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: `%s`", kvp.Key)
			}
		}

		instancesResult = append(instancesResult, instance)

	}

	return instancesResult, nil
}
