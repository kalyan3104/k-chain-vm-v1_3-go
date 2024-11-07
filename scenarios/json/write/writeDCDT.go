package scenjsonwrite

import (
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	oj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/orderedjson"
)

func dcdtTxDataToOJ(dcdtItem *mj.DCDTTxData) *oj.OJsonMap {
	dcdtItemOJ := oj.NewMap()
	if len(dcdtItem.TokenIdentifier.Original) > 0 {
		dcdtItemOJ.Put("tokenIdentifier", bytesFromStringToOJ(dcdtItem.TokenIdentifier))
	}
	if len(dcdtItem.Nonce.Original) > 0 {
		dcdtItemOJ.Put("nonce", uint64ToOJ(dcdtItem.Nonce))
	}
	if len(dcdtItem.Value.Original) > 0 {
		dcdtItemOJ.Put("value", bigIntToOJ(dcdtItem.Value))
	}
	return dcdtItemOJ
}

func dcdtDataToOJ(dcdtItems []*mj.DCDTData) *oj.OJsonMap {
	dcdtItemsOJ := oj.NewMap()
	for _, dcdtItem := range dcdtItems {
		dcdtItemsOJ.Put(dcdtItem.TokenIdentifier.Original, dcdtItemToOJ(dcdtItem))
	}
	return dcdtItemsOJ
}

func dcdtItemToOJ(dcdtItem *mj.DCDTData) oj.OJsonObject {
	if isCompactDCDT(dcdtItem) {
		return bigIntToOJ(dcdtItem.Instances[0].Balance)
	}

	dcdtItemOJ := oj.NewMap()

	// instances
	if len(dcdtItem.Instances) == 1 {
		appendDCDTInstanceToOJ(dcdtItem.Instances[0], dcdtItemOJ)
	} else {
		var convertedList []oj.OJsonObject
		for _, dcdtInstance := range dcdtItem.Instances {
			dcdtInstanceOJ := oj.NewMap()
			appendDCDTInstanceToOJ(dcdtInstance, dcdtInstanceOJ)
			convertedList = append(convertedList, dcdtInstanceOJ)
		}
		instancesOJList := oj.OJsonList(convertedList)
		dcdtItemOJ.Put("instances", &instancesOJList)
	}

	if len(dcdtItem.LastNonce.Original) > 0 {
		dcdtItemOJ.Put("lastNonce", uint64ToOJ(dcdtItem.LastNonce))
	}

	// roles
	if len(dcdtItem.Roles) > 0 {
		var convertedList []oj.OJsonObject
		for _, roleStr := range dcdtItem.Roles {
			convertedList = append(convertedList, &oj.OJsonString{Value: roleStr})
		}
		rolesOJList := oj.OJsonList(convertedList)
		dcdtItemOJ.Put("roles", &rolesOJList)
	}
	if len(dcdtItem.Frozen.Original) > 0 {
		dcdtItemOJ.Put("frozen", uint64ToOJ(dcdtItem.Frozen))
	}

	return dcdtItemOJ
}

func appendDCDTInstanceToOJ(dcdtInstance *mj.DCDTInstance, targetOj *oj.OJsonMap) {
	if len(dcdtInstance.Nonce.Original) > 0 {
		targetOj.Put("nonce", uint64ToOJ(dcdtInstance.Nonce))
	}
	if len(dcdtInstance.Balance.Original) > 0 {
		targetOj.Put("balance", bigIntToOJ(dcdtInstance.Balance))
	}
	if len(dcdtInstance.Creator.Original) > 0 {
		targetOj.Put("creator", bytesFromStringToOJ(dcdtInstance.Creator))
	}
	if len(dcdtInstance.Royalties.Original) > 0 {
		targetOj.Put("royalties", uint64ToOJ(dcdtInstance.Royalties))
	}
	if len(dcdtInstance.Hash.Original) > 0 {
		targetOj.Put("hash", bytesFromStringToOJ(dcdtInstance.Hash))
	}
	if len(dcdtInstance.Uri.Value) > 0 {
		targetOj.Put("uri", bytesFromTreeToOJ(dcdtInstance.Uri))
	}
	if len(dcdtInstance.Attributes.Original) > 0 {
		targetOj.Put("attributes", bytesFromStringToOJ(dcdtInstance.Attributes))
	}
}

func isCompactDCDT(dcdtItem *mj.DCDTData) bool {
	if len(dcdtItem.Instances) != 1 {
		return false
	}
	if len(dcdtItem.Instances[0].Nonce.Original) > 0 {
		return false
	}
	if len(dcdtItem.Roles) > 0 {
		return false
	}
	if len(dcdtItem.Frozen.Original) > 0 {
		return false
	}
	return true
}
