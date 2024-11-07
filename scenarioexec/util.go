package scenarioexec

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-common-go/builtInFunctions"
	worldmock "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/world"
	er "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/expression/reconstructor"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
)

func convertAccount(testAcct *mj.Account, world *worldmock.MockWorld) (*worldmock.Account, error) {
	storage := make(map[string][]byte)
	for _, stkvp := range testAcct.Storage {
		key := string(stkvp.Key.Value)
		storage[key] = stkvp.Value.Value
	}

	if len(testAcct.Address.Value) != 32 {
		return nil, errors.New("bad test: account address should be 32 bytes long")
	}

	account := &worldmock.Account{
		Address:         testAcct.Address.Value,
		Nonce:           testAcct.Nonce.Value,
		Balance:         big.NewInt(0).Set(testAcct.Balance.Value),
		BalanceDelta:    big.NewInt(0),
		DeveloperReward: big.NewInt(0),
		Username:        testAcct.Username.Value,
		Storage:         storage,
		Code:            testAcct.Code.Value,
		OwnerAddress:    testAcct.Owner.Value,
		AsyncCallData:   testAcct.AsyncCallData,
		ShardID:         uint32(testAcct.Shard.Value),
		IsSmartContract: len(testAcct.Code.Value) > 0,
		CodeMetadata: (&vmcommon.CodeMetadata{
			Payable:     true,
			Upgradeable: true,
			Readable:    true,
		}).ToBytes(), // TODO: add explicit fields in scenario JSON
		MockWorld: world,
	}

	for _, scenDCDTData := range testAcct.DCDTData {
		tokenName := scenDCDTData.TokenIdentifier.Value
		isFrozen := scenDCDTData.Frozen.Value > 0
		for _, instance := range scenDCDTData.Instances {
			tokenNonce := instance.Nonce.Value
			tokenKey := worldmock.MakeTokenKey(tokenName, tokenNonce)
			tokenBalance := instance.Balance.Value
			tokenData := &dcdt.DCDigitalToken{
				Value:      tokenBalance,
				Type:       uint32(core.Fungible),
				Properties: makeDCDTUserMetadataBytes(isFrozen),
				TokenMetaData: &dcdt.MetaData{
					Name:       tokenName,
					Nonce:      tokenNonce,
					Creator:    instance.Creator.Value,
					Royalties:  uint32(instance.Royalties.Value),
					Hash:       instance.Hash.Value,
					URIs:       [][]byte{instance.Uri.Value},
					Attributes: instance.Attributes.Value,
				},
			}
			err := account.SetTokenData(tokenKey, tokenData)
			if err != nil {
				return nil, err
			}
			err = account.SetLastNonce(tokenName, scenDCDTData.LastNonce.Value)
			if err != nil {
				return nil, err
			}
		}
		err := account.SetTokenRolesAsStrings(tokenName, scenDCDTData.Roles)
		if err != nil {
			return nil, err
		}
	}

	return account, nil
}

func validateSetStateAccount(scenAccount *mj.Account, converted *worldmock.Account) error {
	err := converted.Validate()
	if err != nil {
		return fmt.Errorf(
			`"setState" step validation failed for account "%s": %w`,
			scenAccount.Address.Original,
			err)
	}
	return nil
}

func makeDCDTUserMetadataBytes(frozen bool) []byte {
	metadata := &builtInFunctions.DCDTUserMetadata{
		Frozen: frozen,
	}

	return metadata.ToBytes()
}

func validateNewAddressMocks(testNAMs []*mj.NewAddressMock) error {
	for _, testNAM := range testNAMs {
		if !worldmock.IsSmartContractAddress(testNAM.NewAddress.Value) {
			return fmt.Errorf(
				`address in "setState" "newAddresses" field should have SC format: %s`,
				testNAM.NewAddress.Original)
		}
	}
	return nil
}

func convertNewAddressMocks(testNAMs []*mj.NewAddressMock) []*worldmock.NewAddressMock {
	var result []*worldmock.NewAddressMock
	for _, testNAM := range testNAMs {
		result = append(result, &worldmock.NewAddressMock{
			CreatorAddress: testNAM.CreatorAddress.Value,
			CreatorNonce:   testNAM.CreatorNonce.Value,
			NewAddress:     testNAM.NewAddress.Value,
		})
	}
	return result
}

func convertBlockInfo(testBlockInfo *mj.BlockInfo) *worldmock.BlockInfo {
	if testBlockInfo == nil {
		return nil
	}

	var randomsSeed [48]byte
	if testBlockInfo.BlockRandomSeed != nil {
		copy(randomsSeed[:], testBlockInfo.BlockRandomSeed.Value)
	}

	result := &worldmock.BlockInfo{
		BlockTimestamp: testBlockInfo.BlockTimestamp.Value,
		BlockNonce:     testBlockInfo.BlockNonce.Value,
		BlockRound:     testBlockInfo.BlockRound.Value,
		BlockEpoch:     uint32(testBlockInfo.BlockEpoch.Value),
		RandomSeed:     &randomsSeed,
	}

	return result
}

// this is a small hack, so we can reuse JSON printing in error messages
func (ae *VMTestExecutor) convertLogToTestFormat(outputLog *vmcommon.LogEntry) *mj.LogEntry {
	testLog := mj.LogEntry{
		Address: mj.JSONCheckBytesReconstructed(
			outputLog.Address,
			ae.exprReconstructor.Reconstruct(outputLog.Address,
				er.AddressHint)),
		Identifier: mj.JSONCheckBytesReconstructed(
			outputLog.Identifier,
			ae.exprReconstructor.Reconstruct(outputLog.Identifier,
				er.StrHint)),
		Data:   mj.JSONCheckBytesReconstructed(outputLog.GetFirstDataItem(), ""),
		Topics: make([]mj.JSONCheckBytes, len(outputLog.Topics)),
	}
	for i, topic := range outputLog.Topics {
		testLog.Topics[i] = mj.JSONCheckBytesReconstructed(topic, "")
	}

	return &testLog
}

func generateTxHash(txIndex string) []byte {
	txIndexBytes := []byte(txIndex)
	if len(txIndexBytes) > 32 {
		return txIndexBytes[:32]
	}
	for i := len(txIndexBytes); i < 32; i++ {
		txIndexBytes = append(txIndexBytes, '.')
	}
	return txIndexBytes
}

func addDCDTToVMInput(dcdtData *mj.DCDTTxData, vmInput *vmcommon.VMInput) {
	if dcdtData != nil {
		vmInput.DCDTTransfers = make([]*vmcommon.DCDTTransfer, 1)
		vmInput.DCDTTransfers[0] = &vmcommon.DCDTTransfer{}
		vmInput.DCDTTransfers[0].DCDTTokenName = dcdtData.TokenIdentifier.Value
		vmInput.DCDTTransfers[0].DCDTValue = dcdtData.Value.Value
		vmInput.DCDTTransfers[0].DCDTTokenNonce = dcdtData.Nonce.Value
		if vmInput.DCDTTransfers[0].DCDTTokenNonce != 0 {
			vmInput.DCDTTransfers[0].DCDTTokenType = uint32(core.NonFungible)
		} else {
			vmInput.DCDTTransfers[0].DCDTTokenType = uint32(core.Fungible)
		}
	}
}
