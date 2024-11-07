package scenjsonwrite

import (
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	oj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/orderedjson"
)

// AccountsToOJ converts a scenarios-format account to an ordered JSON representation.
func AccountsToOJ(accounts []*mj.Account) oj.OJsonObject {
	acctsOJ := oj.NewMap()
	for _, account := range accounts {
		acctOJ := oj.NewMap()
		if len(account.Comment) > 0 {
			acctOJ.Put("comment", stringToOJ(account.Comment))
		}
		acctOJ.Put("nonce", uint64ToOJ(account.Nonce))
		acctOJ.Put("balance", bigIntToOJ(account.Balance))
		if len(account.DCDTData) > 0 {
			acctOJ.Put("dcdt", dcdtDataToOJ(account.DCDTData))
		}
		storageOJ := oj.NewMap()
		for _, st := range account.Storage {
			storageOJ.Put(bytesFromStringToString(st.Key), bytesFromTreeToOJ(st.Value))
		}
		if len(account.Username.Value) > 0 {
			acctOJ.Put("username", bytesFromStringToOJ(account.Username))
		}
		acctOJ.Put("storage", storageOJ)
		acctOJ.Put("code", bytesFromStringToOJ(account.Code))
		if len(account.Owner.Value) > 0 {
			acctOJ.Put("owner", bytesFromStringToOJ(account.Owner))
		}
		if len(account.AsyncCallData) > 0 {
			acctOJ.Put("asyncCallData", stringToOJ(account.AsyncCallData))
		}

		acctsOJ.Put(bytesFromStringToString(account.Address), acctOJ)
	}

	return acctsOJ
}

func checkAccountsToOJ(checkAccounts *mj.CheckAccounts) oj.OJsonObject {
	acctsOJ := oj.NewMap()
	for _, checkAccount := range checkAccounts.Accounts {
		acctOJ := oj.NewMap()
		if len(checkAccount.Comment) > 0 {
			acctOJ.Put("comment", stringToOJ(checkAccount.Comment))
		}
		if !checkAccount.Nonce.IsUnspecified() {
			acctOJ.Put("nonce", checkUint64ToOJ(checkAccount.Nonce))
		}
		if !checkAccount.Balance.IsUnspecified() {
			acctOJ.Put("balance", checkBigIntToOJ(checkAccount.Balance))
		}
		if checkAccount.IgnoreDCDT {
			acctOJ.Put("dcdt", stringToOJ("*"))
		} else {
			if len(checkAccount.CheckDCDTData) > 0 {
				acctOJ.Put("dcdt", checkDCDTDataToOJ(
					checkAccount.CheckDCDTData, checkAccount.MoreDCDTTokensAllowed))
			}
		}
		if !checkAccount.Username.IsUnspecified() {
			acctOJ.Put("username", checkBytesToOJ(checkAccount.Username))
		}
		if checkAccount.IgnoreStorage {
			acctOJ.Put("storage", stringToOJ("*"))
		} else {
			storageOJ := oj.NewMap()
			for _, st := range checkAccount.CheckStorage {
				storageOJ.Put(bytesFromStringToString(st.Key), checkBytesToOJ(st.CheckValue))
			}
			if checkAccount.MoreStorageAllowed {
				storageOJ.Put("+", stringToOJ(""))
			}
			acctOJ.Put("storage", storageOJ)
		}
		if !checkAccount.Code.IsUnspecified() {
			acctOJ.Put("code", checkBytesToOJ(checkAccount.Code))
		}
		if !checkAccount.Owner.IsUnspecified() {
			acctOJ.Put("owner", checkBytesToOJ(checkAccount.Owner))
		}
		if !checkAccount.AsyncCallData.IsUnspecified() {
			acctOJ.Put("asyncCallData", checkBytesToOJ(checkAccount.AsyncCallData))
		}

		acctsOJ.Put(bytesFromStringToString(checkAccount.Address), acctOJ)
	}

	if checkAccounts.MoreAccountsAllowed {
		acctsOJ.Put("+", stringToOJ(""))
	}

	return acctsOJ
}
