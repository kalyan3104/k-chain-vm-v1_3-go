package hostCoretest

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-common-go/builtInFunctions"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/mock"
	worldmock "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/world"
	gasSchedules "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarioexec/gasSchedules"
	testcommon "github.com/kalyan3104/k-chain-vm-v1_3-go/testcommon"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost/hostCore"
	"github.com/stretchr/testify/require"
)

var owner = []byte("owner")
var receiver = []byte("receiver")
var scAddress = []byte("erc20")

func Test_RunERC20Benchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}

	runERC20Benchmark(t, 1000, 4)
}

func runERC20Benchmark(tb testing.TB, nTransfers int, nRuns int) {
	totalTokenSupply := big.NewInt(int64(nTransfers * nRuns))
	host, mockWorld := deploy(tb, totalTokenSupply)

	gasProvided := uint64(5000000000)

	// Prepare ERC20 transfer call input
	transferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: owner,
			Arguments: [][]byte{
				receiver,
				big.NewInt(1).Bytes(),
			},
			CallValue:   big.NewInt(10),
			CallType:    vm.DirectCall,
			GasPrice:    100000000000000,
			GasProvided: gasProvided,
		},
		RecipientAddr: scAddress,
		Function:      "transferToken",
	}

	// Perform ERC20 transfers
	for r := 0; r < nRuns; r++ {
		start := time.Now()
		for i := 0; i < nTransfers; i++ {
			transferInput.GasProvided = gasProvided
			vmOutput, err := host.RunSmartContractCall(transferInput)
			require.Nil(tb, err)
			require.NotNil(tb, vmOutput)
			require.Equal(tb, vmcommon.Ok, vmOutput.ReturnCode)
			require.Equal(tb, "", vmOutput.ReturnMessage)

			_ = mockWorld.UpdateAccounts(vmOutput.OutputAccounts, nil)
		}
		elapsedTime := time.Since(start)
		fmt.Printf("Executing %d ERC20 transfers: %s\n", nTransfers, elapsedTime.String())
	}

	verifyTransfers(tb, mockWorld, totalTokenSupply)
}

func deploy(tb testing.TB, totalTokenSupply *big.Int) (vmhost.VMHost, *worldmock.MockWorld) {
	// Prepare the host
	mockWorld := worldmock.NewMockWorld()
	ownerAccount := &worldmock.Account{
		Address: owner,
		Nonce:   1024,
		Balance: big.NewInt(0),
	}
	mockWorld.AcctMap.PutAccount(ownerAccount)
	mockWorld.NewAddressMocks = append(mockWorld.NewAddressMocks, &worldmock.NewAddressMock{
		CreatorAddress: owner,
		CreatorNonce:   ownerAccount.Nonce,
		NewAddress:     scAddress,
	})

	gasMap, err := gasSchedules.LoadGasScheduleConfig(gasSchedules.GetV2())
	require.Nil(tb, err)

	host, err := hostCore.NewVMHost(mockWorld, &vmhost.VMHostParameters{
		VMType:               testcommon.DefaultVMType,
		BlockGasLimit:        uint64(1000),
		GasSchedule:          gasMap,
		BuiltInFuncContainer: builtInFunctions.NewBuiltInFunctionContainer(),
		ProtectedKeyPrefix:   []byte("E" + "L" + "R" + "O" + "N" + "D"),
		EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == hostCore.SCDeployFlag || flag == hostCore.AheadOfTimeGasUsageFlag || flag == hostCore.RepairCallbackFlag || flag == hostCore.BuiltInFunctionsFlag
			},
		},
	})
	require.Nil(tb, err)

	// Deploy ERC20
	deployInput := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr: owner,
			Arguments: [][]byte{
				totalTokenSupply.Bytes(),
			},
			CallValue:   big.NewInt(0),
			CallType:    vm.DirectCall,
			GasPrice:    0,
			GasProvided: 0xFFFFFFFFFFFFFFFF,
		},
		ContractCode: testcommon.GetTestSCCode("erc20", "../../"),
	}

	ownerAccount.Nonce++ // nonce increases before deploy
	vmOutput, err := host.RunSmartContractCreate(deployInput)
	require.Nil(tb, err)
	require.NotNil(tb, vmOutput)
	require.Equal(tb, "", vmOutput.ReturnMessage)
	require.Equal(tb, vmcommon.Ok, vmOutput.ReturnCode)

	// Ensure the deployment persists in the mock BlockchainHook
	_ = mockWorld.UpdateAccounts(vmOutput.OutputAccounts, nil)
	return host, mockWorld
}

func verifyTransfers(tb testing.TB, mockWorld *worldmock.MockWorld, totalTokenSupply *big.Int) {
	ownerKey := createERC20Key("owner")
	receiverKey := createERC20Key("receiver")

	scStorage := mockWorld.AcctMap.GetAccount(scAddress).Storage
	ownerTokens := big.NewInt(0).SetBytes(scStorage[ownerKey])
	receiverTokens := big.NewInt(0).SetBytes(scStorage[receiverKey])
	require.Equal(tb, vmhost.Zero, ownerTokens)
	require.Equal(tb, totalTokenSupply, receiverTokens)
}

func createERC20Key(accountName string) string {
	keyLength := 32
	key := make([]byte, keyLength)
	key[0] = 1
	key[1] = 0
	i := 2
	for _, c := range accountName {
		key[i] = byte(c)
		i++
		if i == keyLength {
			break
		}
	}
	for q := i; q < keyLength; q++ {
		key[q] = 0
	}

	return string(key)
}
