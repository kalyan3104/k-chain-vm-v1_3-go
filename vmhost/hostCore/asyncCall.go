package hostCore

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-common-go/parsers"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/math"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/wasmer"
)

func (host *vmHost) handleAsyncCallBreakpoint() error {
	log.Trace("async call begin")
	runtime := host.Runtime()
	runtime.SetRuntimeBreakpointValue(vmhost.BreakpointNone)

	asyncCallInfo := runtime.GetAsyncCallInfo()
	execMode, err := host.determineAsyncCallExecutionMode(asyncCallInfo)
	if err != nil {
		log.Trace("async call failed", "error", err)
		return err
	}

	log.Trace("async call", "execMode", execMode)

	if execMode == vmhost.AsyncUnknown {
		err = host.sendAsyncCallToDestination(asyncCallInfo)
		if err != nil {
			log.Trace("async call failed: send cross-shard", "error", err)
		}
		return err
	}

	// Cross-shard calls for built-in functions must be executed in both the
	// sender and destination shards.
	if execMode == vmhost.AsyncBuiltinFuncCrossShard {
		vmOutput, err := host.executeSyncDestinationCall(asyncCallInfo)
		if vmOutput != nil && err != nil {
			log.Trace("async call failed: sync built-in", "error", err,
				"retCode", vmOutput.ReturnCode,
				"message", vmOutput.ReturnMessage)
		}
		return err
	}

	if execMode == vmhost.DCDTTransferOnCallBack {
		// return but keep async call info
		host.outputContext.PrependFinish(asyncCallInfo.Data)
		log.Trace("dcdt transfer on callback")

		// The contract wants to send DCDT back to its original caller
		// via a reversed async call. The reversed async call will not have a
		// callback, therefore the gas locked for callback execution must be
		// restored.
		host.Metering().RestoreGas(asyncCallInfo.GetGasLocked())
		return nil
	}

	// Start calling the destination SC, synchronously.
	destinationVMOutput, destinationErr := host.executeSyncDestinationCall(asyncCallInfo)

	callbackVMOutput, callBackErr := host.executeSyncCallbackCall(asyncCallInfo, destinationVMOutput, destinationErr)

	err = host.processCallbackVMOutput(callbackVMOutput, callBackErr)
	if err != nil {
		return err
	}

	return nil
}

func isDCDTTransferOnReturnDataWithNoAdditionalData(destinationVMOutput *vmcommon.VMOutput) (bool, string, [][]byte) {
	if len(destinationVMOutput.ReturnData) == 0 {
		return false, "", nil
	}

	argParser := parsers.NewCallArgsParser()
	functionName, args, err := argParser.ParseData(string(destinationVMOutput.ReturnData[0]))
	if err != nil {
		return false, "", nil
	}

	return isDCDTTransferOnReturnDataFromFunctionAndArgs(functionName, args)
}

func isDCDTTransferOnReturnDataFromFunctionAndArgs(functionName string, args [][]byte) (bool, string, [][]byte) {

	if functionName == core.BuiltInFunctionDCDTTransfer && len(args) == 2 {
		return true, functionName, args
	}

	if functionName == core.BuiltInFunctionDCDTNFTTransfer && len(args) == 4 {
		return true, functionName, args
	}

	return false, functionName, args
}

func (host *vmHost) determineAsyncCallExecutionMode(asyncCallInfo *vmhost.AsyncCallInfo) (vmhost.AsyncCallExecutionMode, error) {
	runtime := host.Runtime()
	blockchain := host.Blockchain()

	// If ArgParser cannot read the Data field, then this is neither a SC call,
	// nor a built-in function call.
	argParser := parsers.NewCallArgsParser()
	functionName, args, err := argParser.ParseData(string(asyncCallInfo.Data))
	if err != nil {
		return vmhost.AsyncUnknown, err
	}

	sameShard := host.AreInSameShard(runtime.GetSCAddress(), asyncCallInfo.Destination)
	if host.IsBuiltinFunctionName(functionName) {
		if sameShard {
			isDCDTTransfer, _, _ := isDCDTTransferOnReturnDataFromFunctionAndArgs(functionName, args)
			if isDCDTTransfer && runtime.GetVMInput().CallType == vm.AsynchronousCall &&
				bytes.Equal(runtime.GetVMInput().CallerAddr, asyncCallInfo.Destination) {
				return vmhost.DCDTTransferOnCallBack, nil
			}

			return vmhost.AsyncBuiltinFuncIntraShard, nil
		}
		return vmhost.AsyncBuiltinFuncCrossShard, nil
	}

	code, err := blockchain.GetCode(asyncCallInfo.Destination)
	if len(code) > 0 && err == nil {
		return vmhost.SyncCall, nil
	}

	return vmhost.AsyncUnknown, nil
}

func (host *vmHost) executeSyncDestinationCall(asyncCallInfo vmhost.AsyncCallInfoHandler) (*vmcommon.VMOutput, error) {
	destinationCallInput, err := host.createDestinationContractCallInput(asyncCallInfo)
	if err != nil {
		log.Trace("async call: sync dest call failed", "error", err)
		return nil, err
	}

	log.Trace("async call: sync dest call",
		"caller", destinationCallInput.CallerAddr,
		"dest", destinationCallInput.RecipientAddr,
		"func", destinationCallInput.Function,
		"args", destinationCallInput.Arguments)

	destinationVMOutput, _, err := host.ExecuteOnDestContext(destinationCallInput)
	if destinationVMOutput != nil {
		log.Trace("async call: sync dest call",
			"retCode", destinationVMOutput.ReturnCode,
			"message", destinationVMOutput.ReturnMessage,
			"data", destinationVMOutput.ReturnData,
			"error", err)
	}

	return destinationVMOutput, err
}

func (host *vmHost) executeSyncCallbackCall(
	asyncCallInfo vmhost.AsyncCallInfoHandler,
	destinationVMOutput *vmcommon.VMOutput,
	destinationErr error,
) (*vmcommon.VMOutput, error) {
	callbackCallInput, err := host.createCallbackContractCallInput(
		asyncCallInfo,
		destinationVMOutput,
		asyncCallInfo.GetDestination(),
		vmhost.CallbackFunctionName,
		destinationErr,
	)
	if err != nil {
		log.Trace("async call: sync callback failed", "error", err)
		return nil, err
	}

	log.Trace("async call: sync callback",
		"caller", callbackCallInput.CallerAddr,
		"dest", callbackCallInput.RecipientAddr,
		"func", callbackCallInput.Function,
		"args", callbackCallInput.Arguments)

	// Restore gas locked while still on the caller instance; otherwise, the
	// locked gas will appear to have been used twice by the caller instance.
	host.Metering().RestoreGas(asyncCallInfo.GetGasLocked())

	callbackVMOutput, _, callBackErr := host.ExecuteOnDestContext(callbackCallInput)
	if callbackVMOutput != nil {
		log.Trace("async call: sync callback call",
			"retCode", callbackVMOutput.ReturnCode,
			"message", callbackVMOutput.ReturnMessage,
			"data", callbackVMOutput.ReturnData,
			"error", callBackErr)
	}

	return callbackVMOutput, callBackErr
}

func (host *vmHost) canExecuteSynchronously(destination []byte, _ []byte) bool {
	// TODO replace this function in promise-related code below.
	blockchain := host.Blockchain()
	calledSCCode, err := blockchain.GetCode(destination)

	return len(calledSCCode) > 0 && err == nil
}

func (host *vmHost) sendAsyncCallToDestination(asyncCallInfo vmhost.AsyncCallInfoHandler) error {
	runtime := host.Runtime()
	output := host.Output()

	err := output.Transfer(
		asyncCallInfo.GetDestination(),
		runtime.GetSCAddress(),
		asyncCallInfo.GetGasLimit(),
		asyncCallInfo.GetGasLocked(),
		big.NewInt(0).SetBytes(asyncCallInfo.GetValueBytes()),
		asyncCallInfo.GetData(),
		vm.AsynchronousCall,
	)
	if err != nil {
		metering := host.Metering()
		metering.UseGas(metering.GasLeft())
		runtime.FailExecution(err)
		return err
	}

	metering := host.Metering()
	gasLeft := metering.GasLeft()
	metering.UseGas(gasLeft)
	return nil
}

// TODO add locked gas during future refactoring, if needed
func (host *vmHost) sendCallbackToCurrentCaller() error {
	runtime := host.Runtime()
	output := host.Output()
	metering := host.Metering()
	currentCall := runtime.GetVMInput()

	retData := []byte("@" + hex.EncodeToString([]byte(output.ReturnCode().String())))
	for _, data := range output.ReturnData() {
		retData = append(retData, []byte("@"+hex.EncodeToString(data))...)
	}

	err := output.Transfer(
		currentCall.CallerAddr,
		runtime.GetSCAddress(),
		metering.GasLeft(),
		0,
		currentCall.CallValue,
		retData,
		vm.AsynchronousCallBack,
	)
	if err != nil {
		metering.UseGas(metering.GasLeft())
		runtime.FailExecution(err)
		return err
	}

	gasLeft := metering.GasLeft()
	metering.UseGas(gasLeft)
	return nil
}

func (host *vmHost) sendStorageCallbackToDestination(callerAddress, returnData []byte) error {
	runtime := host.Runtime()
	output := host.Output()
	metering := host.Metering()
	currentCall := runtime.GetVMInput()

	err := output.Transfer(
		callerAddress,
		runtime.GetSCAddress(),
		metering.GasLeft(),
		0,
		currentCall.CallValue,
		returnData,
		vm.AsynchronousCallBack,
	)
	if err != nil {
		metering.UseGas(metering.GasLeft())
		runtime.FailExecution(err)
		return err
	}

	return nil
}

func (host *vmHost) createDestinationContractCallInput(asyncCallInfo vmhost.AsyncCallInfoHandler) (*vmcommon.ContractCallInput, error) {
	runtime := host.Runtime()
	sender := runtime.GetSCAddress()
	metering := host.Metering()

	argParser := parsers.NewCallArgsParser()
	function, arguments, err := argParser.ParseData(string(asyncCallInfo.GetData()))
	if err != nil {
		return nil, err
	}

	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     sender,
			Arguments:      arguments,
			CallValue:      big.NewInt(0).SetBytes(asyncCallInfo.GetValueBytes()),
			CallType:       vm.AsynchronousCall,
			GasPrice:       runtime.GetVMInput().GasPrice,
			GasProvided:    metering.GasLeft(),
			GasLocked:      asyncCallInfo.GetGasLocked(),
			CurrentTxHash:  runtime.GetCurrentTxHash(),
			OriginalTxHash: runtime.GetOriginalTxHash(),
		},
		RecipientAddr: asyncCallInfo.GetDestination(),
		Function:      function,
	}

	return contractCallInput, nil
}

func (host *vmHost) computeCallValueFromLastOutputTransfer(destinationVMOutput *vmcommon.VMOutput) *big.Int {
	if !host.IsVMV3Enabled() {
		return big.NewInt(0)
	}
	if len(destinationVMOutput.ReturnData) > 0 {
		return big.NewInt(0)
	}

	returnTransfer := big.NewInt(0)
	callBackReceiver := host.Runtime().GetSCAddress()
	outAcc, ok := destinationVMOutput.OutputAccounts[string(callBackReceiver)]
	if !ok {
		return returnTransfer
	}

	if len(outAcc.OutputTransfers) == 0 {
		return returnTransfer
	}

	lastOutTransfer := outAcc.OutputTransfers[len(outAcc.OutputTransfers)-1]
	if len(lastOutTransfer.Data) == 0 {
		returnTransfer.Set(lastOutTransfer.Value)
	}

	return returnTransfer
}

func (host *vmHost) createCallbackContractCallInput(
	asyncCallInfo vmhost.AsyncCallInfoHandler,
	destinationVMOutput *vmcommon.VMOutput,
	callbackInitiator []byte,
	callbackFunction string,
	destinationErr error,
) (*vmcommon.ContractCallInput, error) {
	metering := host.Metering()
	gasSchedule := metering.GasSchedule()
	runtime := host.Runtime()

	functionName := ""
	isDCDTOnCallBack := false
	dcdtArgs := make([][]byte, 0)
	// always provide return code as the first argument to callback function
	arguments := [][]byte{
		big.NewInt(int64(destinationVMOutput.ReturnCode)).Bytes(),
	}
	returnWithError := false
	if destinationErr == nil && destinationVMOutput.ReturnCode == vmcommon.Ok {
		// when execution went Ok, callBack arguments are:
		// [0, result1, result2, ....]
		isDCDTOnCallBack, functionName, dcdtArgs = isDCDTTransferOnReturnDataWithNoAdditionalData(destinationVMOutput)
		arguments = append(arguments, destinationVMOutput.ReturnData...)
	} else {
		// when execution returned error, callBack arguments are:
		// [error code, error message]
		arguments = append(arguments, []byte(destinationVMOutput.ReturnMessage))
		returnWithError = true
	}

	gasLimit := destinationVMOutput.GasRemaining + asyncCallInfo.GetGasLocked()
	dataLength := host.computeDataLengthFromArguments(callbackFunction, arguments)

	gasToUse := gasSchedule.BaseOpsAPICost.AsyncCallStep
	gas := math.MulUint64(gasSchedule.BaseOperationCost.DataCopyPerByte, uint64(dataLength))
	gasToUse = math.AddUint64(gasToUse, gas)
	if gasLimit <= gasToUse {
		return nil, vmhost.ErrNotEnoughGas
	}
	gasLimit -= gasToUse

	// Return to the sender SC, calling its callback() method.
	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:           callbackInitiator,
			Arguments:            arguments,
			CallValue:            host.computeCallValueFromLastOutputTransfer(destinationVMOutput),
			CallType:             vm.AsynchronousCallBack,
			GasPrice:             runtime.GetVMInput().GasPrice,
			GasProvided:          gasLimit,
			CurrentTxHash:        runtime.GetCurrentTxHash(),
			OriginalTxHash:       runtime.GetOriginalTxHash(),
			ReturnCallAfterError: returnWithError,
		},
		RecipientAddr: runtime.GetSCAddress(),
		Function:      callbackFunction,
	}

	if isDCDTOnCallBack {
		contractCallInput.Function = functionName
		contractCallInput.Arguments = make([][]byte, 0, len(arguments))
		contractCallInput.Arguments = append(contractCallInput.Arguments, dcdtArgs[0], dcdtArgs[1])
		if functionName == core.BuiltInFunctionDCDTNFTTransfer {
			contractCallInput.Arguments = append(contractCallInput.Arguments, dcdtArgs[2], dcdtArgs[3])
		}
		contractCallInput.Arguments = append(contractCallInput.Arguments, []byte(callbackFunction))
		contractCallInput.Arguments = append(contractCallInput.Arguments, big.NewInt(int64(destinationVMOutput.ReturnCode)).Bytes())
		if len(destinationVMOutput.ReturnData) > 1 {
			contractCallInput.Arguments = append(contractCallInput.Arguments, destinationVMOutput.ReturnData[1:]...)
		}
	}

	return contractCallInput, nil
}

func (host *vmHost) processCallbackVMOutput(callbackVMOutput *vmcommon.VMOutput, callBackErr error) error {
	if callBackErr == nil {
		return nil
	}

	runtime := host.Runtime()
	output := host.Output()

	runtime.GetVMInput().GasProvided = 0

	if callbackVMOutput == nil {
		callbackVMOutput = output.CreateVMOutputInCaseOfError(callBackErr)
	}

	output.SetReturnMessage(callbackVMOutput.ReturnMessage)
	output.Finish([]byte(callbackVMOutput.ReturnCode.String()))
	output.Finish(runtime.GetCurrentTxHash())

	return nil
}

func (host *vmHost) computeDataLengthFromArguments(function string, arguments [][]byte) int {
	// Calculate what length would the Data field have, were it of the
	// form "callback@arg1@arg4...

	// TODO this needs tests, especially for the case when the arguments slice
	// contains an empty []byte
	numSeparators := len(arguments)
	dataLength := math.AddUint64(uint64(len(function)), uint64(numSeparators))
	for _, element := range arguments {
		dataLength = math.AddUint64(dataLength, uint64(len(element)))
	}

	return int(dataLength)
}

/**
 * processAsyncInfo takes a list of async calls and for each of them, if the code exists and can be processed on this
 *  host it will. For all others, a vm output account is generated for an actual async call.
 *  Given the fact that the generated async calls that remain pending will be saved on storage, the processing is
 *  done in two steps in order to correctly use all remaining gas. We first split the gas as specified by the developer,
 *  then we save the storage, then we split again the gas to calls that leave this shard.
 *
 * returns a list of pending calls (the ones that should be processed on other hosts)
 */
func (host *vmHost) processAsyncInfo(asyncInfo *vmhost.AsyncContextInfo) (*vmhost.AsyncContextInfo, error) {
	if len(asyncInfo.AsyncContextMap) == 0 {
		return asyncInfo, nil
	}

	err := host.setupAsyncCallsGas(asyncInfo)
	if err != nil {
		return nil, err
	}

	for _, asyncContext := range asyncInfo.AsyncContextMap {
		for _, asyncCall := range asyncContext.AsyncCalls {
			if !host.canExecuteSynchronously(asyncCall.Destination, asyncCall.Data) {
				continue
			}

			procErr := host.processAsyncCall(asyncCall)
			if procErr != nil {
				return nil, procErr
			}
		}
	}

	pendingMapInfo := host.getPendingAsyncCalls(asyncInfo)
	if len(pendingMapInfo.AsyncContextMap) == 0 {
		return pendingMapInfo, nil
	}

	err = host.savePendingAsyncCalls(pendingMapInfo)
	if err != nil {
		return nil, err
	}

	err = host.setupAsyncCallsGas(pendingMapInfo)
	if err != nil {
		return nil, err
	}

	for _, asyncContext := range pendingMapInfo.AsyncContextMap {
		for _, asyncCall := range asyncContext.AsyncCalls {
			if !host.canExecuteSynchronously(asyncCall.Destination, asyncCall.Data) {
				sendErr := host.sendAsyncCallToDestination(asyncCall)
				if sendErr != nil {
					return nil, sendErr
				}
			}
		}
	}

	return pendingMapInfo, nil
}

/**
 * processAsyncCall executes an async call and processes the callback if no extra calls are pending
 */
func (host *vmHost) processAsyncCall(asyncCall *vmhost.AsyncGeneratedCall) error {
	input, _ := host.createDestinationContractCallInput(asyncCall)
	output, asyncMap, executionError := host.ExecuteOnDestContext(input)

	pendingMap := host.getPendingAsyncCalls(asyncMap)
	if len(pendingMap.AsyncContextMap) == 0 {
		return host.callbackAsync(asyncCall, output, executionError)
	}

	return executionError
}

/**
 * callbackAsync will execute a callback from an async call that was ran on this host and set it's status to resolved or rejected
 */
func (host *vmHost) callbackAsync(asyncCall *vmhost.AsyncGeneratedCall, vmOutput *vmcommon.VMOutput, executionError error) error {
	asyncCall.Status = vmhost.AsyncCallResolved
	callbackFunction := asyncCall.SuccessCallback
	if vmOutput.ReturnCode != vmcommon.Ok {
		asyncCall.Status = vmhost.AsyncCallRejected
		callbackFunction = asyncCall.ErrorCallback
	}

	callbackCallInput, err := host.createCallbackContractCallInput(
		asyncCall,
		vmOutput,
		asyncCall.Destination,
		callbackFunction,
		executionError,
	)
	if err != nil {
		return err
	}

	// Callback omits for now any async call - TODO: take into consideration async calls generated from callbacks
	callbackVMOutput, _, callBackErr := host.ExecuteOnDestContext(callbackCallInput)
	err = host.processCallbackVMOutput(callbackVMOutput, callBackErr)
	if err != nil {
		return err
	}

	return nil
}

/**
 * savePendingAsyncCalls takes a list of pending async calls and save them to storage so the info will be available on callback
 */
func (host *vmHost) savePendingAsyncCalls(pendingAsyncMap *vmhost.AsyncContextInfo) error {
	if len(pendingAsyncMap.AsyncContextMap) == 0 {
		return nil
	}

	storage := host.Storage()
	runtime := host.Runtime()

	asyncCallStorageKey := vmhost.CustomStorageKey(vmhost.AsyncDataPrefix, runtime.GetOriginalTxHash())
	data, err := json.Marshal(pendingAsyncMap)
	if err != nil {
		return err
	}

	_, err = storage.SetProtectedStorage(asyncCallStorageKey, data)
	if err != nil {
		return err
	}

	return nil
}

/**
 * saveCrossShardCalls goes through the list of async calls and saves the ones that are cross shard
 */
func (host *vmHost) saveCrossShardCalls(asyncInfo *vmhost.AsyncContextInfo) error {
	crossMap := &vmhost.AsyncContextInfo{
		CallerAddr:      asyncInfo.CallerAddr,
		ReturnData:      asyncInfo.ReturnData,
		AsyncContextMap: make(map[string]*vmhost.AsyncContext),
	}

	for contextIdentifier, asyncContext := range asyncInfo.AsyncContextMap {
		for _, asyncCall := range asyncContext.AsyncCalls {
			if !host.canExecuteSynchronously(asyncCall.Destination, asyncCall.Data) {
				_, ok := crossMap.AsyncContextMap[contextIdentifier]
				if !ok {
					crossMap.AsyncContextMap[contextIdentifier] = &vmhost.AsyncContext{
						Callback:   asyncContext.Callback,
						AsyncCalls: make([]*vmhost.AsyncGeneratedCall, 0),
					}
				}
				crossMap.AsyncContextMap[contextIdentifier].AsyncCalls = append(
					crossMap.AsyncContextMap[contextIdentifier].AsyncCalls,
					asyncCall,
				)
			}
		}
	}

	return host.savePendingAsyncCalls(crossMap)
}

/**
 * getPendingAsyncCalls returns only pending async calls from a list that can also contain resolved/rejected entries
 */
func (host *vmHost) getPendingAsyncCalls(asyncInfo *vmhost.AsyncContextInfo) *vmhost.AsyncContextInfo {
	pendingMap := &vmhost.AsyncContextInfo{
		CallerAddr:      asyncInfo.CallerAddr,
		ReturnData:      asyncInfo.ReturnData,
		AsyncContextMap: make(map[string]*vmhost.AsyncContext),
	}

	for contextIdentifier, asyncContext := range asyncInfo.AsyncContextMap {
		for _, asyncCall := range asyncContext.AsyncCalls {
			if asyncCall.Status != vmhost.AsyncCallPending {
				continue
			}

			_, ok := pendingMap.AsyncContextMap[contextIdentifier]
			if !ok {
				pendingMap.AsyncContextMap[contextIdentifier] = &vmhost.AsyncContext{
					Callback:   asyncContext.Callback,
					AsyncCalls: make([]*vmhost.AsyncGeneratedCall, 0),
				}
			}
			pendingMap.AsyncContextMap[contextIdentifier].AsyncCalls = append(
				pendingMap.AsyncContextMap[contextIdentifier].AsyncCalls,
				asyncCall,
			)
		}
	}

	return pendingMap
}

/**
 * processCallbackStack is triggered when a callback was received from another host through a transaction.
 *  It will return an error if we receive a callback and we don't have it's associated data in the storage.
 *  If the associated callback was found in the pending set, it will be removed - It should not be executed
 *   again since it was executed in the callSCMethod step
 */
func (host *vmHost) processCallbackStack() error {
	runtime := host.Runtime()
	storage := host.Storage()

	storageKey := vmhost.CustomStorageKey(vmhost.AsyncDataPrefix, runtime.GetOriginalTxHash())
	buff := storage.GetStorageUnmetered(storageKey)
	if len(buff) == 0 {
		return nil
	}

	asyncInfo := &vmhost.AsyncContextInfo{}
	err := json.Unmarshal(buff, &asyncInfo)
	if err != nil {
		return err
	}

	vmInput := runtime.GetVMInput()
	var asyncCallPosition int
	var currentContextIdentifier string
	for contextIdentifier, asyncContext := range asyncInfo.AsyncContextMap {
		for position, asyncCall := range asyncContext.AsyncCalls {
			if bytes.Equal(vmInput.CallerAddr, asyncCall.Destination) {
				asyncCallPosition = position
				currentContextIdentifier = contextIdentifier
				break
			}
		}

		if len(currentContextIdentifier) > 0 {
			break
		}
	}

	if len(currentContextIdentifier) == 0 {
		return vmhost.ErrCallBackFuncNotExpected
	}

	// Remove current async call from the pending list
	currentContextCalls := asyncInfo.AsyncContextMap[currentContextIdentifier].AsyncCalls
	contextCallId := len(currentContextCalls) - 1
	if contextCallId >= 0 {
		currentContextCalls[asyncCallPosition] = currentContextCalls[contextCallId]
		currentContextCalls[contextCallId] = nil
		currentContextCalls = currentContextCalls[:contextCallId]
	}

	if len(currentContextCalls) == 0 {
		// call OUR callback for resolving a full context
		delete(asyncInfo.AsyncContextMap, currentContextIdentifier)
	}

	// If we are still waiting for callbacks we return
	if len(asyncInfo.AsyncContextMap) > 0 {
		return nil
	}

	_, err = storage.SetProtectedStorage(storageKey, nil)
	if err != nil {
		return err
	}

	// Now figure out if we can execute the callback here or different shard
	if !host.canExecuteSynchronously(asyncInfo.CallerAddr, asyncInfo.ReturnData) {
		err = host.sendStorageCallbackToDestination(asyncInfo.CallerAddr, asyncInfo.ReturnData)
		if err != nil {
			return err
		}

		return nil
	}

	// The caller is in the same shard, execute it's callback
	// TODO nil pointer exception warning: must refactor
	callbackCallInput, err := host.createCallbackContractCallInput(
		nil,
		host.Output().GetVMOutput(),
		asyncInfo.CallerAddr,
		vmhost.CallbackFunctionName,
		nil,
	)
	if err != nil {
		return err
	}

	callbackVMOutput, _, callBackErr := host.ExecuteOnDestContext(callbackCallInput)
	err = host.processCallbackVMOutput(callbackVMOutput, callBackErr)
	if err != nil {
		return err
	}

	return nil
}

/**
 * setupAsyncCallsGas sets the gasLimit for each async call with the amount of gas provided by the
 *  SC developer. The remaining gas is split between the async calls where the developer
 *  did not specify any gas amount
 */
func (host *vmHost) setupAsyncCallsGas(asyncInfo *vmhost.AsyncContextInfo) error {
	gasLeft := host.Metering().GasLeft()
	gasNeeded := uint64(0)
	callsWithZeroGas := uint64(0)

	for identifier, asyncContext := range asyncInfo.AsyncContextMap {
		for index, asyncCall := range asyncContext.AsyncCalls {
			var err error
			gasNeeded, err = math.AddUint64WithErr(gasNeeded, asyncCall.ProvidedGas)
			if err != nil {
				return err
			}

			if gasNeeded > gasLeft {
				return vmhost.ErrNotEnoughGas
			}

			if asyncCall.ProvidedGas == 0 {
				callsWithZeroGas++
				continue
			}

			asyncInfo.AsyncContextMap[identifier].AsyncCalls[index].GasLimit = asyncCall.ProvidedGas
		}
	}

	if callsWithZeroGas == 0 {
		return nil
	}

	if gasLeft <= gasNeeded {
		return vmhost.ErrNotEnoughGas
	}

	gasShare := (gasLeft - gasNeeded) / callsWithZeroGas
	for identifier, asyncContext := range asyncInfo.AsyncContextMap {
		for index, asyncCall := range asyncContext.AsyncCalls {
			if asyncCall.ProvidedGas == 0 {
				asyncInfo.AsyncContextMap[identifier].AsyncCalls[index].GasLimit = gasShare
			}
		}
	}

	return nil
}

func (host *vmHost) getFunctionByCallType(callType vm.CallType) (wasmer.ExportedFunctionCallback, error) {
	runtime := host.Runtime()

	if callType != vm.AsynchronousCallBack {
		return runtime.GetFunctionToCall()
	}

	asyncInfo, err := host.getCurrentAsyncInfo()
	if err != nil {
		return nil, err
	}

	vmInput := runtime.GetVMInput()

	customCallback := false
	for _, asyncContext := range asyncInfo.AsyncContextMap {
		for _, asyncCall := range asyncContext.AsyncCalls {
			if bytes.Equal(vmInput.CallerAddr, asyncCall.Destination) {
				customCallback = true
				runtime.SetCustomCallFunction(asyncCall.SuccessCallback)
				break
			}
		}

		if customCallback {
			break
		}
	}

	function, err := runtime.GetFunctionToCall()
	if err != nil && !customCallback {
		log.Trace("get function by call type", "error", vmhost.ErrNilCallbackFunction)
		return nil, vmhost.ErrNilCallbackFunction
	}

	return function, nil
}

func (host *vmHost) getCurrentAsyncInfo() (*vmhost.AsyncContextInfo, error) {
	runtime := host.Runtime()
	storage := host.Storage()

	asyncInfo := &vmhost.AsyncContextInfo{}
	storageKey := vmhost.CustomStorageKey(vmhost.AsyncDataPrefix, runtime.GetOriginalTxHash())
	buff := storage.GetStorageUnmetered(storageKey)
	if len(buff) == 0 {
		return asyncInfo, nil
	}

	err := json.Unmarshal(buff, &asyncInfo)
	if err != nil {
		return nil, err
	}

	return asyncInfo, nil
}
