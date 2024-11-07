package scenarioexec

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	worldmock "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/world"
	er "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/expression/reconstructor"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	oj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/orderedjson"
)

// ExecuteCheckStateStep executes a CheckStateStep defined by the current scenario.
func (ae *VMTestExecutor) ExecuteCheckStateStep(step *mj.CheckStateStep) error {
	if len(step.Comment) > 0 {
		log.Trace("CheckStateStep", "comment", step.Comment)
	}

	return ae.checkAccounts(step.CheckAccounts)
}

func (ae *VMTestExecutor) checkAccounts(checkAccounts *mj.CheckAccounts) error {
	if !checkAccounts.MoreAccountsAllowed {
		for worldAcctAddr := range ae.World.AcctMap {
			postAcctMatch := mj.FindCheckAccount(checkAccounts.Accounts, []byte(worldAcctAddr))
			if postAcctMatch == nil && !bytes.Equal(vmcommon.SystemAccountAddress, []byte(worldAcctAddr)) {
				return fmt.Errorf("unexpected account address: %s",
					ae.exprReconstructor.Reconstruct(
						[]byte(worldAcctAddr),
						er.AddressHint))
			}
		}
	}

	for _, expectedAcct := range checkAccounts.Accounts {
		matchingAcct, isMatch := ae.World.AcctMap[string(expectedAcct.Address.Value)]
		if !isMatch {
			return fmt.Errorf("account %s expected but not found after running test",
				expectedAcct.Address.Original)
		}

		if !bytes.Equal(matchingAcct.Address, expectedAcct.Address.Value) {
			return fmt.Errorf("bad account address %s",
				ae.exprReconstructor.Reconstruct(
					matchingAcct.Address,
					er.AddressHint))
		}

		if !expectedAcct.Nonce.Check(matchingAcct.Nonce) {
			return fmt.Errorf("bad account nonce. Account: %s. Want: \"%s\". Have: \"%d\"",
				expectedAcct.Address.Original,
				expectedAcct.Nonce.Original,
				matchingAcct.Nonce)
		}

		if !expectedAcct.Balance.Check(matchingAcct.Balance) {
			return fmt.Errorf("bad account balance. Account: %s. Want: \"%s\". Have: \"%s\"",
				expectedAcct.Address.Original,
				expectedAcct.Balance.Original,
				ae.exprReconstructor.ReconstructFromBigInt(matchingAcct.Balance))
		}

		if !expectedAcct.Username.Check(matchingAcct.Username) {
			return fmt.Errorf("bad account username. Account: %s. Want: %s. Have: \"%s\"",
				expectedAcct.Address.Original,
				oj.JSONString(expectedAcct.Username.Original),
				ae.exprReconstructor.Reconstruct(
					matchingAcct.Username,
					er.StrHint))
		}

		if !expectedAcct.Code.Check(matchingAcct.Code) {
			return fmt.Errorf("bad account code. Account: %s. Want: %s. Have: \"%s\"",
				expectedAcct.Address.Original,
				oj.JSONString(expectedAcct.Code.Original),
				ae.exprReconstructor.Reconstruct(
					matchingAcct.Code,
					er.CodeHint))
		}

		// currently ignoring asyncCallData that is unspecified in the json
		if !expectedAcct.AsyncCallData.IsUnspecified() &&
			!expectedAcct.AsyncCallData.Check([]byte(matchingAcct.AsyncCallData)) {
			return fmt.Errorf("bad async call data. Account: %s. Want: [%s]. Have: [%s]",
				expectedAcct.Address.Original,
				expectedAcct.AsyncCallData.Original,
				matchingAcct.AsyncCallData)
		}

		err := ae.checkAccountStorage(expectedAcct, matchingAcct)
		if err != nil {
			return err
		}

		err = ae.checkAccountDCDT(expectedAcct, matchingAcct)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ae *VMTestExecutor) checkAccountStorage(expectedAcct *mj.CheckAccount, matchingAcct *worldmock.Account) error {
	if expectedAcct.IgnoreStorage {
		return nil
	}

	expectedStorage := make(map[string]mj.JSONCheckBytes)
	for _, stkvp := range expectedAcct.CheckStorage {
		expectedStorage[string(stkvp.Key.Value)] = stkvp.CheckValue
	}

	allKeys := make(map[string]bool)
	for k := range expectedStorage {
		allKeys[k] = true
	}
	for k := range matchingAcct.Storage {
		allKeys[k] = true
	}
	storageError := ""
	for k := range allKeys {
		// ignore all reserved keys
		if strings.HasPrefix(k, core.ProtectedKeyPrefix) {
			continue
		}

		want, specified := expectedStorage[k]
		if !specified {
			if expectedAcct.MoreStorageAllowed {
				// if `"+": ""` was written in the test, any unspecified entries are allowed,
				// which is equivalent to treating them all as "*".
				want = mj.JSONCheckBytesStar()
			} else {
				// otherwise, by default, any unexpected storage key leads to a test failure
				want = mj.JSONCheckBytesUnspecified()
			}
		}
		have := matchingAcct.StorageValue(k)

		if !want.Check(have) {
			storageError += fmt.Sprintf(
				"\n  for key %s: Want: %s. Have: \"%s\"",
				ae.exprReconstructor.Reconstruct([]byte(k), er.NoHint),
				oj.JSONString(want.Original),
				ae.exprReconstructor.Reconstruct(have, er.NoHint))
		}
	}
	if len(storageError) > 0 {
		return fmt.Errorf("wrong account storage for account \"%s\":%s",
			expectedAcct.Address.Original, storageError)
	}
	return nil
}

func (ae *VMTestExecutor) checkAccountDCDT(expectedAcct *mj.CheckAccount, matchingAcct *worldmock.Account) error {
	if expectedAcct.IgnoreDCDT {
		return nil
	}

	accountAddress := expectedAcct.Address.Original
	expectedTokens := getExpectedTokens(expectedAcct)
	accountTokens, err := matchingAcct.GetFullMockDCDTData()
	if err != nil {
		return err
	}

	allTokenNames := make(map[string]bool)
	for tokenName := range expectedTokens {
		allTokenNames[tokenName] = true
	}
	for tokenName := range accountTokens {
		allTokenNames[tokenName] = true
	}
	var errors []error
	for tokenName := range allTokenNames {
		expectedToken := expectedTokens[tokenName]
		accountToken := accountTokens[tokenName]
		if expectedToken == nil {
			expectedToken = &mj.CheckDCDTData{
				TokenIdentifier: mj.JSONBytesFromString{
					Value:    []byte(tokenName),
					Original: ae.exprReconstructor.Reconstruct([]byte(tokenName), er.StrHint),
				},
				Instances: []*mj.CheckDCDTInstance{},
				LastNonce: mj.JSONCheckUint64{Value: 0, Original: ""},
				Roles:     []string{},
			}
		} else if accountToken == nil {
			accountToken = &worldmock.MockDCDTData{
				TokenIdentifier: []byte(tokenName),
				Instances:       []*dcdt.DCDigitalToken{},
				LastNonce:       0,
				Roles:           [][]byte{},
			}
		}

		errors = append(errors, ae.checkTokenState(accountAddress, tokenName, expectedToken, accountToken)...)
	}

	errorString := makeErrorString(errors)
	if len(errorString) > 0 {
		return fmt.Errorf("mismatch for account \"%s\":%s", accountAddress, errorString)
	}

	return nil
}

func getExpectedTokens(expectedAcct *mj.CheckAccount) map[string]*mj.CheckDCDTData {
	expectedTokens := make(map[string]*mj.CheckDCDTData)
	for _, expectedTokenData := range expectedAcct.CheckDCDTData {
		tokenName := expectedTokenData.TokenIdentifier.Value
		expectedTokens[string(tokenName)] = expectedTokenData
	}

	return expectedTokens
}

func (ae *VMTestExecutor) checkTokenState(
	accountAddress string,
	tokenName string,
	expectedToken *mj.CheckDCDTData,
	accountToken *worldmock.MockDCDTData) []error {

	var errors []error

	errors = append(errors, ae.checkTokenInstances(accountAddress, tokenName, expectedToken, accountToken)...)

	if !expectedToken.LastNonce.Check(accountToken.LastNonce) {
		errors = append(errors, fmt.Errorf("bad account DCDT last nonce. Account: %s. Token: %s. Want: \"%s\". Have: %d",
			accountAddress,
			tokenName,
			expectedToken.LastNonce.Original,
			accountToken.LastNonce))
	}

	errors = append(errors, checkTokenRoles(accountAddress, tokenName, expectedToken, accountToken)...)

	return errors
}

func (ae *VMTestExecutor) checkTokenInstances(
	accountAddress string,
	tokenName string,
	expectedToken *mj.CheckDCDTData,
	accountToken *worldmock.MockDCDTData) []error {

	var errors []error

	allNonces := make(map[uint64]bool)
	expectedInstances := make(map[uint64]*mj.CheckDCDTInstance)
	accountInstances := make(map[uint64]*dcdt.DCDigitalToken)
	for _, expectedInstance := range expectedToken.Instances {
		nonce := expectedInstance.Nonce.Value
		allNonces[nonce] = true
		expectedInstances[nonce] = expectedInstance
	}
	for _, accountInstance := range accountToken.Instances {
		nonce := accountInstance.TokenMetaData.Nonce
		allNonces[nonce] = true
		accountInstances[nonce] = accountInstance
	}

	for nonce := range allNonces {
		expectedInstance := expectedInstances[nonce]
		accountInstance := accountInstances[nonce]

		if expectedInstance == nil {
			expectedInstance = &mj.CheckDCDTInstance{
				Nonce:   mj.JSONCheckUint64{Value: nonce, Original: ""},
				Balance: mj.JSONCheckBigInt{Value: big.NewInt(0), Original: ""},
			}
		} else if accountInstance == nil {
			accountInstance = &dcdt.DCDigitalToken{
				Value: big.NewInt(0),
				TokenMetaData: &dcdt.MetaData{
					Name:  []byte(tokenName),
					Nonce: nonce,
				},
			}
		}

		if !expectedInstance.Balance.Check(accountInstance.Value) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad balance. Want: \"%s\". Have: \"%d\"",
				tokenName,
				nonce,
				expectedInstance.Balance.Original,
				accountInstance.Value))
		}
		if !expectedInstance.Creator.IsUnspecified() &&
			!expectedInstance.Creator.Check(accountInstance.TokenMetaData.Creator) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad creator. Want: %s. Have: \"%s\"",
				tokenName,
				nonce,
				oj.JSONString(expectedInstance.Creator.Original),
				ae.exprReconstructor.Reconstruct(
					accountInstance.TokenMetaData.Creator,
					er.AddressHint)))
		}
		if !expectedInstance.Royalties.IsUnspecified() &&
			!expectedInstance.Royalties.Check(uint64(accountInstance.TokenMetaData.Royalties)) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad royalties. Want: \"%s\". Have: \"%s\"",
				tokenName,
				nonce,
				expectedInstance.Royalties.Original,
				ae.exprReconstructor.ReconstructFromUint64(
					uint64(accountInstance.TokenMetaData.Royalties))))
		}
		if !expectedInstance.Hash.IsUnspecified() &&
			!expectedInstance.Hash.Check(accountInstance.TokenMetaData.Hash) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad hash. Want: %s. Have: %s",
				tokenName,
				nonce,
				oj.JSONString(expectedInstance.Hash.Original),
				ae.exprReconstructor.Reconstruct(
					accountInstance.TokenMetaData.Hash,
					er.NoHint)))
		}
		if len(accountInstance.TokenMetaData.URIs) > 1 {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: More than one URI currently not supported",
				tokenName,
				nonce))
		}
		var actualUri []byte
		if len(accountInstance.TokenMetaData.URIs) == 1 {
			actualUri = accountInstance.TokenMetaData.URIs[0]
		}
		if !expectedInstance.Uri.IsUnspecified() &&
			!expectedInstance.Uri.Check(actualUri) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad URI. Want: %s. Have: \"%s\"",
				tokenName,
				nonce,
				oj.JSONString(expectedInstance.Uri.Original),
				ae.exprReconstructor.Reconstruct(
					actualUri,
					er.StrHint)))
		}
		if !expectedInstance.Attributes.IsUnspecified() &&
			!expectedInstance.Attributes.Check(accountInstance.TokenMetaData.Attributes) {
			errors = append(errors, fmt.Errorf(
				"for token: %s, nonce: %d: Bad attributes. Want: %s. Have: \"%s\"",
				tokenName,
				nonce,
				oj.JSONString(expectedInstance.Attributes.Original),
				ae.exprReconstructor.Reconstruct(
					accountInstance.TokenMetaData.Attributes,
					er.StrHint)))
		}

	}

	return errors
}

func checkTokenRoles(
	accountAddress string,
	tokenName string,
	expectedToken *mj.CheckDCDTData,
	accountToken *worldmock.MockDCDTData) []error {

	var errors []error

	allRoles := make(map[string]bool)
	expectedRoles := make(map[string]bool)
	accountRoles := make(map[string]bool)

	for _, expectedRole := range expectedToken.Roles {
		allRoles[expectedRole] = true
		expectedRoles[expectedRole] = true
	}
	for _, accountRole := range accountToken.Roles {
		allRoles[string(accountRole)] = true
		accountRoles[string(accountRole)] = true
	}
	for role := range allRoles {
		if !expectedRoles[role] {
			errors = append(errors, fmt.Errorf("unexpected DCDT role. Account: %s. Token: %s. Role: %s",
				accountAddress,
				tokenName,
				role))
		}
		if !accountRoles[role] {
			errors = append(errors, fmt.Errorf("missing DCDT role. Account: %s. Token: %s. Role: %s",
				accountAddress,
				tokenName,
				role))
		}
	}

	return errors
}

func makeErrorString(errors []error) string {
	errorString := ""
	for _, err := range errors {
		errorString += "\n  " + err.Error()
	}
	return errorString
}
