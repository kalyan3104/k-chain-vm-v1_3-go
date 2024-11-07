package contracts

import (
	"fmt"
	"math/big"

	"github.com/kalyan3104/k-chain-vm-common-go/txDataBuilder"
	mock "github.com/kalyan3104/k-chain-vm-v1_3-go/mock/context"
	test "github.com/kalyan3104/k-chain-vm-v1_3-go/testcommon"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost/vmhooks"
)

// ExecDCDTTransferAndCallChild is an exposed mock contract method
func ExecDCDTTransferAndCallChild(instanceMock *mock.InstanceMock, config interface{}) {
	testConfig := config.(DirectCallGasTestConfig)
	instanceMock.AddMockMethod("execDCDTTransferAndCall", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		host.Metering().UseGas(testConfig.GasUsedByParent)

		arguments := host.Runtime().Arguments()
		if len(arguments) != 3 {
			host.Runtime().SignalUserError("need 3 arguments")
			return instance
		}

		input := test.DefaultTestContractCallInput()
		input.CallerAddr = host.Runtime().GetSCAddress()
		input.GasProvided = testConfig.GasProvidedToChild
		input.Arguments = [][]byte{
			test.DCDTTestTokenName,
			big.NewInt(int64(testConfig.DCDTTokensToTransfer)).Bytes(),
			arguments[2],
		}
		input.RecipientAddr = arguments[0]
		input.Function = string(arguments[1])

		returnValue := ExecuteOnDestContextInMockContracts(host, input)
		if returnValue != 0 {
			host.Runtime().FailExecution(fmt.Errorf("Return value %d", returnValue))
		}

		return instance
	})
}

// ExecDCDTTransferWithAPICall is an exposed mock contract method
func ExecDCDTTransferWithAPICall(instanceMock *mock.InstanceMock, config interface{}) {
	testConfig := config.(DirectCallGasTestConfig)
	instanceMock.AddMockMethod("execDCDTTransferWithAPICall", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		host.Metering().UseGas(testConfig.GasUsedByParent)

		arguments := host.Runtime().Arguments()
		if len(arguments) != 3 {
			host.Runtime().SignalUserError("need 3 arguments")
			return instance
		}

		input := test.DefaultTestContractCallInput()
		input.CallerAddr = host.Runtime().GetSCAddress()
		input.GasProvided = testConfig.GasProvidedToChild
		input.Arguments = [][]byte{
			test.DCDTTestTokenName,
			big.NewInt(int64(testConfig.DCDTTokensToTransfer)).Bytes(),
			arguments[2],
		}
		input.RecipientAddr = arguments[0]

		functionName := arguments[1]
		args := [][]byte{arguments[2]}

		vmhooks.TransferDCDTNFTExecuteWithTypedArgs(
			host,
			big.NewInt(int64(testConfig.DCDTTokensToTransfer)),
			test.DCDTTestTokenName,
			input.RecipientAddr,
			0,
			int64(testConfig.GasProvidedToChild),
			[]byte(functionName),
			args)

		return instance
	})
}

// ExecDCDTTransferAndAsyncCallChild is an exposed mock contract method
func ExecDCDTTransferAndAsyncCallChild(instanceMock *mock.InstanceMock, config interface{}) {
	testConfig := config.(*AsyncCallTestConfig)
	instanceMock.AddMockMethod("execDCDTTransferAndAsyncCall", func() *mock.InstanceMock {
		host := instanceMock.Host
		instance := mock.GetMockInstance(host)
		host.Metering().UseGas(testConfig.GasUsedByParent)

		arguments := host.Runtime().Arguments()
		if len(arguments) != 3 {
			host.Runtime().SignalUserError("need 3 arguments")
			return instance
		}

		functionToCallOnChild := arguments[2]

		receiver := arguments[0]
		builtInFunction := arguments[1]

		callData := txDataBuilder.NewBuilder()
		// function to be called on child
		callData.Func(string(builtInFunction))
		callData.Bytes(test.DCDTTestTokenName)
		callData.Bytes(big.NewInt(int64(testConfig.DCDTTokensToTransfer)).Bytes())
		callData.Bytes(functionToCallOnChild)

		value := big.NewInt(0).Bytes()

		err := host.Runtime().ExecuteAsyncCall(receiver, callData.ToBytes(), value)

		if err != nil {
			host.Runtime().FailExecution(err)
		}

		return instance
	})
}
