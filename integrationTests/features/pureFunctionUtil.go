package featuresintegrationtest

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"

	"github.com/kalyan3104/k-chain-core-go/core"
	vmi "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-common-go/builtInFunctions"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/config"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/mock"
	worldhook "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/world"
	mj "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/json/model"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost/hostCore"
	"github.com/stretchr/testify/require"
)

type pureFunctionIO struct {
	functionName    string
	arguments       [][]byte
	expectedStatus  vmi.ReturnCode
	expectedMessage string
	expectedResults [][]byte
}

var testVMType = []byte{0, 0}

type resultInterpreter func([]byte) *big.Int
type logProgress func(testCaseIndex, testCaseCount int)

type pureFunctionExecutor struct {
	world           *worldhook.MockWorld
	vm              vmi.VMExecutionHandler
	contractAddress []byte
	userAddress     []byte
}

func newPureFunctionExecutor() (*pureFunctionExecutor, error) {
	world := worldhook.NewMockWorld()

	blockGasLimit := uint64(10000000)
	gasSchedule := config.MakeGasMapForTests()
	vm, err := hostCore.NewVMHost(world, &vmhost.VMHostParameters{
		VMType:               testVMType,
		BlockGasLimit:        blockGasLimit,
		GasSchedule:          gasSchedule,
		BuiltInFuncContainer: builtInFunctions.NewBuiltInFunctionContainer(),
		ProtectedKeyPrefix:   []byte("E" + "L" + "R" + "O" + "N" + "D"),
		EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == hostCore.SCDeployFlag || flag == hostCore.AheadOfTimeGasUsageFlag || flag == hostCore.RepairCallbackFlag || flag == hostCore.BuiltInFunctionsFlag
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &pureFunctionExecutor{
		world: world,
		vm:    vm,
	}, nil
}

func (pfe *pureFunctionExecutor) initAccounts(contractPath string) {
	pfe.contractAddress = []byte("contract_addr_________________s1")
	pfe.userAddress = []byte("user_addr_____________________s1")

	scCode, err := ioutil.ReadFile(contractPath)
	if err != nil {
		panic(err)
	}

	pfe.world.AcctMap.PutAccount(&worldhook.Account{
		Address: pfe.contractAddress,
		Nonce:   0,
		Balance: big.NewInt(0),
		Storage: make(map[string][]byte),
		Code:    scCode,
	})

	pfe.world.AcctMap.PutAccount(&worldhook.Account{
		Address: pfe.userAddress,
		Nonce:   0,
		Balance: big.NewInt(0x100000000),
		Storage: make(map[string][]byte),
		Code:    []byte{},
	})
}

func (pfe *pureFunctionExecutor) scCall(testCase *pureFunctionIO) (*vmi.VMOutput, error) {
	input := &vmi.ContractCallInput{
		RecipientAddr: pfe.contractAddress,
		Function:      testCase.functionName,
		VMInput: vmi.VMInput{
			CallerAddr:  pfe.userAddress,
			Arguments:   testCase.arguments,
			CallValue:   big.NewInt(0),
			GasPrice:    1,
			GasProvided: 100000000,
		},
	}

	return pfe.vm.RunSmartContractCall(input)
}

func (pfe *pureFunctionExecutor) checkTxResults(
	testCase *pureFunctionIO,
	output *vmi.VMOutput,
	resultInterpreter resultInterpreter) error {

	if output.ReturnCode != testCase.expectedStatus {
		return fmt.Errorf("result code mismatch. Want: %d. Have: %d (%s). Message: %s",
			int(testCase.expectedStatus), int(output.ReturnCode), output.ReturnCode.String(), output.ReturnMessage)
	}

	if output.ReturnMessage != testCase.expectedMessage {
		return fmt.Errorf("result message mismatch. Want: %s. Have: %s",
			testCase.expectedMessage, output.ReturnMessage)
	}

	// check result
	if len(output.ReturnData) != len(testCase.expectedResults) {
		return fmt.Errorf("result length mismatch. Want: %s. Have: %s",
			mj.ResultAsString(testCase.expectedResults), mj.ResultAsString(output.ReturnData))
	}
	for i, expected := range testCase.expectedResults {
		wantNum := resultInterpreter(expected)
		haveNum := resultInterpreter(output.ReturnData[i])
		if wantNum.Cmp(haveNum) != 0 {
			var argStr []string
			for _, arg := range testCase.arguments {
				argNum := resultInterpreter(arg)
				argStr = append(argStr, fmt.Sprintf("%d", argNum))
			}
			return fmt.Errorf("result mismatch. Want: %d. Have: %d. Call: %s(%s)",
				wantNum, haveNum, testCase.functionName, strings.Join(argStr, ", "))
		}
	}

	return nil
}

func (pfe *pureFunctionExecutor) executePureFunctionTests(t *testing.T,
	testCases []*pureFunctionIO,
	resultInterpreter resultInterpreter,
	logProgress logProgress) {

	// RUN!
	for testCaseIndex, testCase := range testCases {
		if logProgress != nil {
			logProgress(testCaseIndex, len(testCases))
		}

		output, err := pfe.scCall(testCase)
		require.Nil(t, err)

		err = pfe.checkTxResults(testCase, output, resultInterpreter)
		require.Nil(t, err)
	}
}
