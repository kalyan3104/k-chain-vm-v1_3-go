package worldmock

import (
	"fmt"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
)

// DCDTTokenKeyPrefix is the prefix of storage keys belonging to DCDT tokens.
var DCDTTokenKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTKeyIdentifier)

// DCDTRoleKeyPrefix is the prefix of storage keys belonging to DCDT roles.
var DCDTRoleKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTRoleIdentifier + core.DCDTKeyIdentifier)

// DCDTNonceKeyPrefix is the prefix of storage keys belonging to DCDT nonces.
var DCDTNonceKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTNFTLatestNonceIdentifier)

// GetTokenBalance returns the DCDT balance of an account for the given token
// key (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenBalance(address []byte, tokenKey []byte) (*big.Int, error) {
	account := bf.World.AcctMap.GetAccount(address)
	return account.GetTokenBalance(tokenKey)
}

// SetTokenBalance sets the DCDT balance of an account for the given token
// key (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) SetTokenBalance(address []byte, tokenKey []byte, balance *big.Int) error {
	account := bf.World.AcctMap.GetAccount(address)
	return account.SetTokenBalance(tokenKey, balance)
}

// GetTokenData gets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenData(address []byte, tokenKey []byte) (*dcdt.DCDigitalToken, error) {
	account := bf.World.AcctMap.GetAccount(address)
	return account.GetTokenData(tokenKey)
}

// SetTokenData sets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) SetTokenData(address []byte, tokenKey []byte, tokenData *dcdt.DCDigitalToken) error {
	account := bf.World.AcctMap.GetAccount(address)
	return account.SetTokenData(tokenKey, tokenData)
}

// PerformDirectDCDTTransfer calls the real DCDTTransfer function immediately;
// only works for in-shard transfers for now, but it will be expanded to
// cross-shard.
// TODO rewrite to simulate what the SCProcessor does when executing a tx with
// data "DCDTTransfer@token@value@contractfunc@contractargs..."
// TODO this function duplicates code from host.ExecuteDCDTTransfer(), must refactor
func (bf *BuiltinFunctionsWrapper) PerformDirectDCDTTransfer(
	sender []byte,
	receiver []byte,
	token []byte,
	nonce uint64,
	value *big.Int,
	callType vm.CallType,
	gasLimit uint64,
	gasPrice uint64,
) (uint64, error) {
	dcdtTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    callType,
			GasPrice:    gasPrice,
			GasProvided: gasLimit,
			GasLocked:   0,
		},
		RecipientAddr:     receiver,
		Function:          core.BuiltInFunctionDCDTTransfer,
		AllowInitFunction: false,
	}

	if nonce > 0 {
		dcdtTransferInput.Function = core.BuiltInFunctionDCDTNFTTransfer
		dcdtTransferInput.RecipientAddr = dcdtTransferInput.CallerAddr
		nonceAsBytes := big.NewInt(0).SetUint64(nonce).Bytes()
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, nonceAsBytes, value.Bytes(), receiver)
	} else {
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, value.Bytes())
	}

	vmOutput, err := bf.ProcessBuiltInFunction(dcdtTransferInput)
	if err != nil {
		return 0, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return 0, fmt.Errorf(
			"DCDTtransfer failed: retcode = %d, msg = %s",
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage)
	}

	return vmOutput.GasRemaining, nil
}
