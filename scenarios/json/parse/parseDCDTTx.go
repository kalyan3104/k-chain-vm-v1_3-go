package scenjsonparse

import (
	"errors"
	"fmt"

	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	oj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/orderedjson"
)

func (p *Parser) processTxDCDT(txDcdtRaw oj.OJsonObject) (*mj.DCDTTxData, error) {
	fieldMap, isMap := txDcdtRaw.(*oj.OJsonMap)
	if !isMap {
		return nil, errors.New("unmarshalled account object is not a map")
	}

	dcdtData := mj.DCDTTxData{}
	var err error

	for _, kvp := range fieldMap.OrderedKV {
		switch kvp.Key {
		case "tokenIdentifier":
			dcdtData.TokenIdentifier, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid DCDT token name: %w", err)
			}
		case "nonce":
			dcdtData.Nonce, err = p.processUint64(kvp.Value)
			if err != nil {
				return nil, errors.New("invalid account nonce")
			}
		case "value":
			dcdtData.Value, err = p.processBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid DCDT balance: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown transaction DCDT data field: %s", kvp.Key)
		}
	}

	return &dcdtData, nil
}
