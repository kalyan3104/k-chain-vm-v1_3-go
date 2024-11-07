package common

import (
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
)

type SerializableVMOutput struct {
	ReturnData              [][]byte
	ReturnCode              vmcommon.ReturnCode
	ReturnMessage           string
	GasRemaining            uint64
	GasRefund               *big.Int
	CorrectedOutputAccounts []*SerializableOutputAccount
	DeletedAccounts         [][]byte
	TouchedAccounts         [][]byte
	Logs                    []*vmcommon.LogEntry
}

func NewSerializableVMOutput(vmOutput *vmcommon.VMOutput) *SerializableVMOutput {
	if vmOutput == nil {
		return &SerializableVMOutput{}
	}

	o := &SerializableVMOutput{
		ReturnData:              vmOutput.ReturnData,
		ReturnCode:              vmOutput.ReturnCode,
		ReturnMessage:           vmOutput.ReturnMessage,
		GasRemaining:            vmOutput.GasRemaining,
		GasRefund:               vmOutput.GasRefund,
		CorrectedOutputAccounts: make([]*SerializableOutputAccount, 0, len(vmOutput.OutputAccounts)),
		DeletedAccounts:         vmOutput.DeletedAccounts,
		TouchedAccounts:         vmOutput.TouchedAccounts,
		Logs:                    vmOutput.Logs,
	}

	for _, account := range vmOutput.OutputAccounts {
		o.CorrectedOutputAccounts = append(o.CorrectedOutputAccounts, NewSerializableOutputAccount(account))
	}

	return o
}

func (o *SerializableVMOutput) ConvertToVMOutput() *vmcommon.VMOutput {
	accountsMap := make(map[string]*vmcommon.OutputAccount)

	for _, item := range o.CorrectedOutputAccounts {
		accountsMap[string(item.Address)] = item.ConvertToOutputAccount()
	}

	return &vmcommon.VMOutput{
		ReturnData:      o.ReturnData,
		ReturnCode:      o.ReturnCode,
		ReturnMessage:   o.ReturnMessage,
		GasRemaining:    o.GasRemaining,
		GasRefund:       o.GasRefund,
		OutputAccounts:  accountsMap,
		DeletedAccounts: o.DeletedAccounts,
		TouchedAccounts: o.TouchedAccounts,
		Logs:            o.Logs,
	}
}

type SerializableOutputAccount struct {
	Address             []byte
	Nonce               uint64
	Balance             *big.Int
	BalanceDelta        *big.Int
	StorageUpdates      []*vmcommon.StorageUpdate
	Code                []byte
	CodeMetadata        []byte
	GasUsed             uint64
	Transfers           []SerializableOutputTransfer
	CodeDeployerAddress []byte
}

type SerializableOutputTransfer struct {
	Value         *big.Int
	Data          []byte
	GasLimit      uint64
	GasLocked     uint64
	CallType      vm.CallType
	SenderAddress []byte
}

func NewSerializableOutputAccount(account *vmcommon.OutputAccount) *SerializableOutputAccount {
	a := &SerializableOutputAccount{
		Address:             account.Address,
		Nonce:               account.Nonce,
		Balance:             account.Balance,
		BalanceDelta:        account.BalanceDelta,
		StorageUpdates:      make([]*vmcommon.StorageUpdate, 0, len(account.StorageUpdates)),
		Code:                account.Code,
		CodeMetadata:        account.CodeMetadata,
		GasUsed:             account.GasUsed,
		CodeDeployerAddress: account.CodeDeployerAddress,
	}

	a.Transfers = make([]SerializableOutputTransfer, len(account.OutputTransfers))
	for i, transfer := range account.OutputTransfers {
		serializableTransfer := SerializableOutputTransfer{
			Value:         transfer.Value,
			Data:          transfer.Data,
			GasLimit:      transfer.GasLimit,
			GasLocked:     transfer.GasLocked,
			CallType:      transfer.CallType,
			SenderAddress: transfer.SenderAddress,
		}
		a.Transfers[i] = serializableTransfer
	}

	for _, storageUpdate := range account.StorageUpdates {
		a.StorageUpdates = append(a.StorageUpdates, storageUpdate)
	}

	return a
}

func (a *SerializableOutputAccount) ConvertToOutputAccount() *vmcommon.OutputAccount {
	updatesMap := make(map[string]*vmcommon.StorageUpdate)

	for _, item := range a.StorageUpdates {
		updatesMap[string(item.Offset)] = item
	}

	outAcc := &vmcommon.OutputAccount{
		Address:             a.Address,
		Nonce:               a.Nonce,
		Balance:             a.Balance,
		BalanceDelta:        a.BalanceDelta,
		StorageUpdates:      updatesMap,
		Code:                a.Code,
		CodeMetadata:        a.CodeMetadata,
		GasUsed:             a.GasUsed,
		CodeDeployerAddress: a.CodeDeployerAddress,
	}
	outAcc.OutputTransfers = make([]vmcommon.OutputTransfer, len(a.Transfers))
	for i, transfer := range a.Transfers {
		outPutTransfer := vmcommon.OutputTransfer{
			Value:         transfer.Value,
			GasLimit:      transfer.GasLimit,
			GasLocked:     transfer.GasLocked,
			Data:          transfer.Data,
			CallType:      transfer.CallType,
			SenderAddress: transfer.SenderAddress,
		}
		outAcc.OutputTransfers[i] = outPutTransfer
	}

	return outAcc
}
