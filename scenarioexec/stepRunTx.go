package scenarioexec

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
)

func (ae *VMTestExecutor) executeTx(txIndex string, tx *mj.Transaction) (*vmcommon.VMOutput, error) {
	ae.World.CreateStateBackup()

	var err error
	defer func() {
		if err != nil {
			errRollback := ae.World.RollbackChanges()
			if errRollback != nil {
				err = errRollback
			}
		} else {
			errCommit := ae.World.CommitChanges()
			if errCommit != nil {
				err = errCommit
			}
		}
	}()

	gasForExecution := uint64(0)

	if tx.Type.HasSender() {
		beforeErr := ae.World.UpdateWorldStateBefore(
			tx.From.Value,
			tx.GasLimit.Value,
			tx.GasPrice.Value)
		if beforeErr != nil {
			err = fmt.Errorf("could not set up tx %s: %w", txIndex, beforeErr)
			return nil, err
		}

		gasForExecution = tx.GasLimit.Value
		if tx.DCDTValue != nil {
			gasRemaining, err := ae.directDCDTTransferFromTx(tx)
			if err != nil {
				return nil, err
			}

			gasForExecution = gasRemaining
		}
	}

	// we also use fake vm outputs for transactions that don't use the VM, just for convenience
	var output *vmcommon.VMOutput

	if !ae.senderHasEnoughBalance(tx) {
		// out of funds is handled by the protocol, so it needs to be mocked here
		output = outOfFundsResult()
	} else {
		switch tx.Type {
		case mj.ScDeploy:
			output, err = ae.scCreate(txIndex, tx, gasForExecution)
			if err != nil {
				return nil, err
			}
		case mj.ScQuery:
			// imitates the behaviour of the protocol
			// the sender is the contract itself during SC queries
			tx.From = tx.To
			// gas restrictions waived during SC queries
			tx.GasLimit.Value = math.MaxUint64
			gasForExecution = math.MaxUint64
			fallthrough
		case mj.ScCall:
			output, err = ae.scCall(txIndex, tx, gasForExecution)
			if err != nil {
				return nil, err
			}
		case mj.Transfer:
			output = ae.simpleTransferOutput(tx)
		case mj.ValidatorReward:
			output, err = ae.validatorRewardOutput(tx)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unknown transaction type")
		}
	}

	if output.ReturnCode == vmcommon.Ok {
		err := ae.updateStateAfterTx(tx, output)
		if err != nil {
			return nil, err
		}
	} else {
		err = fmt.Errorf(
			"tx step failed: retcode=%d, msg=%s",
			output.ReturnCode, output.ReturnMessage)
	}

	return output, nil
}

func (ae *VMTestExecutor) senderHasEnoughBalance(tx *mj.Transaction) bool {
	if !tx.Type.HasSender() {
		return true
	}
	sender := ae.World.AcctMap.GetAccount(tx.From.Value)
	return sender.Balance.Cmp(tx.Value.Value) >= 0
}

func (ae *VMTestExecutor) simpleTransferOutput(tx *mj.Transaction) *vmcommon.VMOutput {
	outputAccounts := make(map[string]*vmcommon.OutputAccount)
	outputAccounts[string(tx.To.Value)] = &vmcommon.OutputAccount{
		Address:      tx.To.Value,
		BalanceDelta: tx.Value.Value,
	}

	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  outputAccounts,
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}
}

func (ae *VMTestExecutor) validatorRewardOutput(tx *mj.Transaction) (*vmcommon.VMOutput, error) {
	reward := tx.Value.Value
	recipient := ae.World.AcctMap.GetAccount(tx.To.Value)
	if recipient == nil {
		return nil, fmt.Errorf("tx recipient (address: %s) does not exist", hex.EncodeToString(tx.To.Value))
	}
	recipient.BalanceDelta = reward
	storageReward := recipient.StorageValue(RewardKey)
	storageReward = big.NewInt(0).Add(
		big.NewInt(0).SetBytes(storageReward),
		reward).Bytes()

	outputAccounts := make(map[string]*vmcommon.OutputAccount)
	outputAccounts[string(tx.To.Value)] = &vmcommon.OutputAccount{
		Address:      tx.To.Value,
		BalanceDelta: tx.Value.Value,
		StorageUpdates: map[string]*vmcommon.StorageUpdate{
			RewardKey: {
				Offset: []byte(RewardKey),
				Data:   storageReward,
			},
		},
	}

	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  outputAccounts,
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}, nil
}

func outOfFundsResult() *vmcommon.VMOutput {
	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.OutOfFunds,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  make(map[string]*vmcommon.OutputAccount),
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}
}

func (ae *VMTestExecutor) scCreate(txIndex string, tx *mj.Transaction, gasLimit uint64) (*vmcommon.VMOutput, error) {
	txHash := generateTxHash(txIndex)
	vmInput := vmcommon.VMInput{
		CallerAddr:     tx.From.Value,
		Arguments:      mj.JSONBytesFromTreeValues(tx.Arguments),
		CallValue:      tx.Value.Value,
		GasPrice:       tx.GasPrice.Value,
		GasProvided:    gasLimit,
		OriginalTxHash: txHash,
		CurrentTxHash:  txHash,
		DCDTTransfers:  make([]*vmcommon.DCDTTransfer, 0),
	}
	addDCDTToVMInput(tx.DCDTValue, &vmInput)
	input := &vmcommon.ContractCreateInput{
		ContractCode: tx.Code.Value,
		VMInput:      vmInput,
	}

	return ae.vm.RunSmartContractCreate(input)
}

func (ae *VMTestExecutor) scCall(txIndex string, tx *mj.Transaction, gasLimit uint64) (*vmcommon.VMOutput, error) {
	recipient := ae.World.AcctMap.GetAccount(tx.To.Value)
	if recipient == nil {
		return nil, fmt.Errorf("tx recipient (address: %s) does not exist", hex.EncodeToString(tx.To.Value))
	}
	if len(recipient.Code) == 0 {
		return nil, fmt.Errorf("tx recipient (address: %s) is not a smart contract", hex.EncodeToString(tx.To.Value))
	}
	txHash := generateTxHash(txIndex)
	vmInput := vmcommon.VMInput{
		CallerAddr:     tx.From.Value,
		Arguments:      mj.JSONBytesFromTreeValues(tx.Arguments),
		CallValue:      tx.Value.Value,
		GasPrice:       tx.GasPrice.Value,
		GasProvided:    gasLimit,
		OriginalTxHash: txHash,
		CurrentTxHash:  txHash,
		DCDTTransfers:  make([]*vmcommon.DCDTTransfer, 0),
	}
	addDCDTToVMInput(tx.DCDTValue, &vmInput)
	input := &vmcommon.ContractCallInput{
		RecipientAddr: tx.To.Value,
		Function:      tx.Function,
		VMInput:       vmInput,
	}

	return ae.vm.RunSmartContractCall(input)
}

func (ae *VMTestExecutor) directDCDTTransferFromTx(tx *mj.Transaction) (uint64, error) {
	return ae.World.BuiltinFuncs.PerformDirectDCDTTransfer(
		tx.From.Value,
		tx.To.Value,
		tx.DCDTValue.TokenIdentifier.Value,
		tx.DCDTValue.Nonce.Value,
		tx.DCDTValue.Value.Value,
		vm.DirectCall,
		tx.GasLimit.Value,
		tx.GasPrice.Value)
}

func (ae *VMTestExecutor) updateStateAfterTx(
	tx *mj.Transaction,
	output *vmcommon.VMOutput) error {

	// subtract call value from sender (this is not reflected in the delta)
	// except for validatorReward, there is no sender there
	if tx.Type.HasSender() {
		_ = ae.World.UpdateBalanceWithDelta(tx.From.Value, big.NewInt(0).Neg(tx.Value.Value))
	}

	// update accounts based on deltas
	updErr := ae.World.UpdateAccounts(output.OutputAccounts, output.DeletedAccounts)
	if updErr != nil {
		return updErr
	}

	// sum of all balance deltas should equal call value (unless we got an error)
	// (unless it is validatorReward, when funds just pop into existence)
	if tx.Type.HasSender() {
		sumOfBalanceDeltas := big.NewInt(0)
		for _, oa := range output.OutputAccounts {
			sumOfBalanceDeltas = sumOfBalanceDeltas.Add(sumOfBalanceDeltas, oa.BalanceDelta)
		}
		if sumOfBalanceDeltas.Cmp(tx.Value.Value) != 0 {
			return fmt.Errorf("sum of balance deltas should equal call value. Sum of balance deltas: %d (0x%x). Call value: %d (0x%x)",
				sumOfBalanceDeltas, sumOfBalanceDeltas, tx.Value.Value, tx.Value.Value)
		}
	}

	return nil
}
