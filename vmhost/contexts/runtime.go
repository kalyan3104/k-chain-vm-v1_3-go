package contexts

import (
	"bytes"
	"errors"
	"fmt"
	builtinMath "math"
	"math/big"
	"unsafe"

	logger "github.com/kalyan3104/k-chain-logger-go"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/math"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/wasmer"
)

var logRuntime = logger.GetOrCreate("vm/runtime")

var _ vmhost.RuntimeContext = (*runtimeContext)(nil)

// Defined as a constant here, not present in gasSchedule V1, V2, V3
const MaxMemoryGrow = uint64(10)
const MaxMemoryGrowDelta = uint64(10)

type runtimeContext struct {
	host         vmhost.VMHost
	instance     wasmer.InstanceHandler
	vmInput      *vmcommon.VMInput
	scAddress    []byte
	codeSize     uint64
	callFunction string
	vmType       []byte
	readOnly     bool

	verifyCode bool

	stateStack    []*runtimeContext
	instanceStack []wasmer.InstanceHandler

	maxWasmerInstances uint64

	asyncCallInfo    *vmhost.AsyncCallInfo
	asyncContextInfo *vmhost.AsyncContextInfo

	validator *wasmValidator

	useWarmInstance     bool
	warmInstanceAddress []byte
	warmInstance        wasmer.InstanceHandler

	instanceBuilder vmhost.InstanceBuilder

	errors vmhost.WrappableError
}

// NewRuntimeContext creates a new runtimeContext
func NewRuntimeContext(
	host vmhost.VMHost,
	vmType []byte,
	useWarmInstance bool,
	builtInFuncContainer vmcommon.BuiltInFunctionContainer,
) (*runtimeContext, error) {
	scAPINames := host.GetAPIMethods().Names()

	context := &runtimeContext{
		host:                host,
		vmType:              vmType,
		stateStack:          make([]*runtimeContext, 0),
		instanceStack:       make([]wasmer.InstanceHandler, 0),
		validator:           newWASMValidator(scAPINames, builtInFuncContainer),
		useWarmInstance:     useWarmInstance,
		warmInstanceAddress: nil,
		warmInstance:        nil,
		errors:              nil,
	}

	context.instanceBuilder = &wasmerInstanceBuilder{}
	context.InitState()

	return context, nil
}

// InitState initializes all the contexts fields with default data.
func (context *runtimeContext) InitState() {
	context.vmInput = &vmcommon.VMInput{}
	context.scAddress = make([]byte, 0)
	context.callFunction = ""
	context.verifyCode = false
	context.readOnly = false
	context.asyncCallInfo = nil
	context.asyncContextInfo = &vmhost.AsyncContextInfo{
		AsyncContextMap: make(map[string]*vmhost.AsyncContext),
	}
	context.errors = nil

	logRuntime.Trace("init state")
}

// ReplaceInstanceBuilder replaces the instance builder, allowing the creation
// of mocked Wasmer instances
// TODO remove after implementing proper mocking of
// Wasmer instances; this is used for tests only
func (context *runtimeContext) ReplaceInstanceBuilder(builder vmhost.InstanceBuilder) {
	context.instanceBuilder = builder
}

func (context *runtimeContext) setWarmInstanceWhenNeeded(gasLimit uint64) bool {
	scAddress := context.GetSCAddress()
	useWarm := context.useWarmInstance && context.warmInstanceAddress != nil && bytes.Equal(scAddress, context.warmInstanceAddress)
	if scAddress != nil && useWarm {
		logRuntime.Trace("reusing warm instance")

		context.instance = context.warmInstance
		context.SetPointsUsed(0)
		context.instance.SetGasLimit(gasLimit)

		context.SetRuntimeBreakpointValue(vmhost.BreakpointNone)
		return true
	}

	return false
}

// StartWasmerInstance creates a new wasmer instance if the maxWasmerInstances has not been reached.
func (context *runtimeContext) StartWasmerInstance(contract []byte, gasLimit uint64, newCode bool) error {
	if context.RunningInstancesCount() >= context.maxWasmerInstances {
		context.instance = nil
		logRuntime.Error("create instance", "error", vmhost.ErrMaxInstancesReached)
		return vmhost.ErrMaxInstancesReached
	}

	warmInstanceUsed := context.setWarmInstanceWhenNeeded(gasLimit)
	if warmInstanceUsed {
		return nil
	}

	blockchain := context.host.Blockchain()
	codeHash := blockchain.GetCodeHash(context.GetSCAddress())
	compiledCodeUsed := context.makeInstanceFromCompiledCode(codeHash, gasLimit, newCode)
	if compiledCodeUsed {
		return nil
	}

	return context.makeInstanceFromContractByteCode(contract, codeHash, gasLimit, newCode)
}

func (context *runtimeContext) makeInstanceFromCompiledCode(codeHash []byte, gasLimit uint64, newCode bool) bool {
	if !context.host.IsAheadOfTimeCompileEnabled() {
		return false
	}

	if newCode || len(codeHash) == 0 {
		return false
	}

	blockchain := context.host.Blockchain()
	found, compiledCode := blockchain.GetCompiledCode(codeHash)
	if !found {
		logRuntime.Trace("instance creation", "code", "cached compilation", "error", "compiled code was not found")
		return false
	}

	gasSchedule := context.host.Metering().GasSchedule()
	options := wasmer.CompilationOptions{
		GasLimit:           gasLimit,
		UnmeteredLocals:    uint64(gasSchedule.WASMOpcodeCost.LocalsUnmetered),
		MaxMemoryGrow:      MaxMemoryGrow,
		MaxMemoryGrowDelta: MaxMemoryGrowDelta,
		OpcodeTrace:        false,
		Metering:           true,
		RuntimeBreakpoints: true,
	}
	newInstance, err := context.instanceBuilder.NewInstanceFromCompiledCodeWithOptions(compiledCode, options)
	if err != nil {
		logRuntime.Error("instance creation", "code", "cached compilation", "error", err)
		return false
	}

	context.instance = newInstance

	hostReference := uintptr(unsafe.Pointer(&context.host))
	context.instance.SetContextData(hostReference)
	context.verifyCode = false

	logRuntime.Trace("new instance created", "code", "cached compilation")
	return true
}

func (context *runtimeContext) makeInstanceFromContractByteCode(contract []byte, codeHash []byte, gasLimit uint64, newCode bool) error {
	gasSchedule := context.host.Metering().GasSchedule()
	options := wasmer.CompilationOptions{
		GasLimit:           gasLimit,
		UnmeteredLocals:    uint64(gasSchedule.WASMOpcodeCost.LocalsUnmetered),
		MaxMemoryGrow:      MaxMemoryGrow,
		MaxMemoryGrowDelta: MaxMemoryGrowDelta,
		OpcodeTrace:        false,
		Metering:           true,
		RuntimeBreakpoints: true,
	}
	newInstance, err := context.instanceBuilder.NewInstanceWithOptions(contract, options)
	if err != nil {
		context.instance = nil
		logRuntime.Trace("instance creation", "code", "bytecode", "error", err)
		return err
	}

	context.instance = newInstance

	if newCode || len(codeHash) == 0 {
		codeHash, err = context.host.Crypto().Sha256(contract)
		if err != nil {
			context.CleanWasmerInstance()
			logRuntime.Error("instance creation", "code", "bytecode", "error", err)
			return err
		}
	}

	context.saveCompiledCode(codeHash)

	hostReference := uintptr(unsafe.Pointer(&context.host))
	context.instance.SetContextData(hostReference)

	if newCode {
		err = context.VerifyContractCode()
		if err != nil {
			context.CleanWasmerInstance()
			logRuntime.Trace("instance creation", "code", "bytecode", "error", err)
			return err
		}
	}

	if context.useWarmInstance {
		context.warmInstanceAddress = context.GetSCAddress()
		context.warmInstance = context.instance
		logRuntime.Trace("updated warm instance")
	}

	logRuntime.Trace("new instance created", "code", "bytecode")

	return nil
}

// GetSCCode returns the SC code of the current SC.
func (context *runtimeContext) GetSCCode() ([]byte, error) {
	blockchain := context.host.Blockchain()
	code, err := blockchain.GetCode(context.scAddress)
	if err != nil {
		return nil, err
	}

	context.codeSize = uint64(len(code))
	return code, nil
}

// GetSCCodeSize returns the size of the current SC code.
func (context *runtimeContext) GetSCCodeSize() uint64 {
	return context.codeSize
}

func (context *runtimeContext) saveCompiledCode(codeHash []byte) {
	compiledCode, err := context.instance.Cache()
	if err != nil {
		logRuntime.Error("getCompiledCode from instance", "error", err)
		return
	}

	blockchain := context.host.Blockchain()
	blockchain.SaveCompiledCode(codeHash, compiledCode)
}

// IsWarmInstance returns true if there is a warm instance equal to the current wasmer instance.
func (context *runtimeContext) IsWarmInstance() bool {
	if context.instance != nil && context.instance == context.warmInstance {
		return true
	}

	return false
}

// ResetWarmInstance clears the fields for the current wasmer instance, warm instance, and warm instance address
func (context *runtimeContext) ResetWarmInstance() {
	if context.instance == nil {
		return
	}

	context.instance.Clean()

	context.instance = nil
	context.warmInstanceAddress = nil
	context.warmInstance = nil
	logRuntime.Trace("warm instance cleaned")
}

// MustVerifyNextContractCode sets the verifyCode field to true
func (context *runtimeContext) MustVerifyNextContractCode() {
	context.verifyCode = true
}

// SetMaxInstanceCount sets the maxWasmerInstances field to the given value
func (context *runtimeContext) SetMaxInstanceCount(maxInstances uint64) {
	context.maxWasmerInstances = maxInstances
}

// InitStateFromContractCallInput initializes the runtime context state with the values from the given input
func (context *runtimeContext) InitStateFromContractCallInput(input *vmcommon.ContractCallInput) {
	context.SetVMInput(&input.VMInput)
	context.scAddress = input.RecipientAddr
	context.callFunction = input.Function
	// Reset async map for initial state
	context.asyncContextInfo = &vmhost.AsyncContextInfo{
		CallerAddr:      input.CallerAddr,
		AsyncContextMap: make(map[string]*vmhost.AsyncContext),
	}

	logRuntime.Trace("init state from call input",
		"caller", input.CallerAddr,
		"contract", input.RecipientAddr,
		"func", input.Function,
		"args", input.Arguments)
}

// SetCustomCallFunction sets the given string as the callFunction field.
func (context *runtimeContext) SetCustomCallFunction(callFunction string) {
	context.callFunction = callFunction
	logRuntime.Trace("set custom call function", "function", callFunction)
}

// PushState appends the current runtime state to the state stack; this
// includes the currently running Wasmer instance.
func (context *runtimeContext) PushState() {
	newState := &runtimeContext{
		scAddress:        context.scAddress,
		callFunction:     context.callFunction,
		readOnly:         context.readOnly,
		asyncCallInfo:    context.asyncCallInfo,
		asyncContextInfo: context.asyncContextInfo,
	}
	newState.SetVMInput(context.vmInput)

	context.stateStack = append(context.stateStack, newState)

	// Also preserve the currently running Wasmer instance at the top of the
	// instance stack; when the corresponding call to popInstance() is made, a
	// check is made to ensure that the running instance will not be cleaned
	// while still required for execution.
	context.pushInstance()
}

// PopSetActiveState removes the latest entry from the state stack and sets it as the current
// runtime context state.
func (context *runtimeContext) PopSetActiveState() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	prevState := context.stateStack[stateStackLen-1]
	context.stateStack = context.stateStack[:stateStackLen-1]

	context.SetVMInput(prevState.vmInput)
	context.scAddress = prevState.scAddress
	context.callFunction = prevState.callFunction
	context.readOnly = prevState.readOnly
	context.asyncCallInfo = prevState.asyncCallInfo
	context.asyncContextInfo = prevState.asyncContextInfo
	context.popInstance()
}

// PopDiscard removes the latest entry from the state stack
func (context *runtimeContext) PopDiscard() {
	stateStackLen := len(context.stateStack)
	if stateStackLen == 0 {
		return
	}

	context.stateStack = context.stateStack[:stateStackLen-1]
	context.popInstance()
}

// ClearStateStack reinitializes the state stack.
func (context *runtimeContext) ClearStateStack() {
	context.stateStack = make([]*runtimeContext, 0)
}

// pushInstance appends the current wasmer instance to the instance stack.
func (context *runtimeContext) pushInstance() {
	context.instanceStack = append(context.instanceStack, context.instance)
}

// popInstance removes the latest entry from the wasmer instance stack and sets it
// as the current wasmer instance
func (context *runtimeContext) popInstance() {
	instanceStackLen := len(context.instanceStack)
	if instanceStackLen == 0 {
		return
	}

	prevInstance := context.instanceStack[instanceStackLen-1]
	context.instanceStack = context.instanceStack[:instanceStackLen-1]

	if prevInstance == context.instance {
		// The current Wasmer instance was previously pushed on the instance stack,
		// but a new Wasmer instance has not been created in the meantime. This
		// means that the instance at the top of the stack is the same as the
		// current instance, so it cannot be cleaned, because the execution will
		// resume on it. Popping will therefore only remove the top of the stack,
		// without cleaning anything.
		return
	}

	context.CleanWasmerInstance()
	context.instance = prevInstance
}

// RunningInstancesCount returns the length of the instance stack.
func (context *runtimeContext) RunningInstancesCount() uint64 {
	return uint64(len(context.instanceStack))
}

// GetVMType returns the vm type for the current context.
func (context *runtimeContext) GetVMType() []byte {
	return context.vmType
}

// GetVMInput returns the vm input for the current context.
func (context *runtimeContext) GetVMInput() *vmcommon.VMInput {
	return context.vmInput
}

func copyDCDTTransfer(dcdtTransfer *vmcommon.DCDTTransfer) *vmcommon.DCDTTransfer {
	newDCDTTransfer := &vmcommon.DCDTTransfer{
		DCDTValue:      big.NewInt(0).Set(dcdtTransfer.DCDTValue),
		DCDTTokenType:  dcdtTransfer.DCDTTokenType,
		DCDTTokenNonce: dcdtTransfer.DCDTTokenNonce,
		DCDTTokenName:  make([]byte, len(dcdtTransfer.DCDTTokenName)),
	}
	copy(newDCDTTransfer.DCDTTokenName, dcdtTransfer.DCDTTokenName)
	return newDCDTTransfer
}

// SetVMInput sets the given vm input as the current context vm input.
func (context *runtimeContext) SetVMInput(vmInput *vmcommon.VMInput) {
	if vmInput == nil {
		context.vmInput = vmInput
		return
	}

	context.vmInput = &vmcommon.VMInput{
		CallType:             vmInput.CallType,
		GasPrice:             vmInput.GasPrice,
		GasProvided:          vmInput.GasProvided,
		GasLocked:            vmInput.GasLocked,
		CallValue:            big.NewInt(0),
		ReturnCallAfterError: vmInput.ReturnCallAfterError,
	}

	if vmInput.CallValue != nil {
		context.vmInput.CallValue.Set(vmInput.CallValue)
	}

	if len(vmInput.CallerAddr) > 0 {
		context.vmInput.CallerAddr = make([]byte, len(vmInput.CallerAddr))
		copy(context.vmInput.CallerAddr, vmInput.CallerAddr)
	}

	context.vmInput.DCDTTransfers = make([]*vmcommon.DCDTTransfer, len(vmInput.DCDTTransfers))

	if len(vmInput.DCDTTransfers) > 0 {
		for i, dcdtTransfer := range vmInput.DCDTTransfers {
			context.vmInput.DCDTTransfers[i] = copyDCDTTransfer(dcdtTransfer)
		}
	}

	if len(vmInput.OriginalTxHash) > 0 {
		context.vmInput.OriginalTxHash = make([]byte, len(vmInput.OriginalTxHash))
		copy(context.vmInput.OriginalTxHash, vmInput.OriginalTxHash)
	}

	if len(vmInput.CurrentTxHash) > 0 {
		context.vmInput.CurrentTxHash = make([]byte, len(vmInput.CurrentTxHash))
		copy(context.vmInput.CurrentTxHash, vmInput.CurrentTxHash)
	}

	if len(vmInput.Arguments) > 0 {
		context.vmInput.Arguments = make([][]byte, len(vmInput.Arguments))
		for i, arg := range vmInput.Arguments {
			context.vmInput.Arguments[i] = make([]byte, len(arg))
			copy(context.vmInput.Arguments[i], arg)
		}
	}
}

// GetSCAddress returns the SC address from the current context.
func (context *runtimeContext) GetSCAddress() []byte {
	return context.scAddress
}

// SetSCAddress sets the given address as the scAddress for the current context.
func (context *runtimeContext) SetSCAddress(scAddress []byte) {
	context.scAddress = scAddress
}

// GetCurrentTxHash returns the txHash from the vmInput of the current context.
func (context *runtimeContext) GetCurrentTxHash() []byte {
	return context.vmInput.CurrentTxHash
}

// GetOriginalTxHash returns the originalTxHash from the vmInput of the current context.
func (context *runtimeContext) GetOriginalTxHash() []byte {
	return context.vmInput.OriginalTxHash
}

// Function returns the callFunction for the current context.
func (context *runtimeContext) Function() string {
	return context.callFunction
}

// Arguments returns the arguments from the vmInput of the current context.
func (context *runtimeContext) Arguments() [][]byte {
	return context.vmInput.Arguments
}

// ExtractCodeUpgradeFromArgs extracts the arguments needed for a code upgrade from the vmInput.
func (context *runtimeContext) ExtractCodeUpgradeFromArgs() ([]byte, []byte, error) {
	const numMinUpgradeArguments = 2

	arguments := context.vmInput.Arguments
	if len(arguments) < numMinUpgradeArguments {
		return nil, nil, vmhost.ErrInvalidUpgradeArguments
	}

	code := arguments[0]
	codeMetadata := arguments[1]
	context.vmInput.Arguments = context.vmInput.Arguments[numMinUpgradeArguments:]
	return code, codeMetadata, nil
}

// FailExecution sets the returnMessage, returnCode and runtimeBreakpoint according to the given error.
func (context *runtimeContext) FailExecution(err error) {
	context.host.Output().SetReturnCode(vmcommon.ExecutionFailed)

	var message string
	if err != nil {
		message = err.Error()
		context.AddError(err)
	} else {
		message = "execution failed"
		context.AddError(errors.New(message))
	}

	context.host.Output().SetReturnMessage(message)
	context.SetRuntimeBreakpointValue(vmhost.BreakpointExecutionFailed)

	traceMessage := message
	if err != nil {
		traceMessage = err.Error()
	}
	logRuntime.Trace("execution failed", "message", traceMessage)
}

// SignalUserError sets the returnMessage, returnCode and runtimeBreakpoint according an user error.
func (context *runtimeContext) SignalUserError(message string) {
	context.host.Output().SetReturnCode(vmcommon.UserError)
	context.host.Output().SetReturnMessage(message)
	context.SetRuntimeBreakpointValue(vmhost.BreakpointSignalError)
	context.AddError(errors.New(message))
	logRuntime.Trace("user error signalled", "message", message)
}

// SetRuntimeBreakpointValue sets the given value as a breakpoint value.
func (context *runtimeContext) SetRuntimeBreakpointValue(value vmhost.BreakpointValue) {
	context.instance.SetBreakpointValue(uint64(value))
	logRuntime.Trace("runtime breakpoint set", "breakpoint", value)
}

// GetRuntimeBreakpointValue returns the breakpoint value for the current wasmer instance.
func (context *runtimeContext) GetRuntimeBreakpointValue() vmhost.BreakpointValue {
	return vmhost.BreakpointValue(context.instance.GetBreakpointValue())
}

// VerifyContractCode checks the current wasmer instance for enough memory and for correct functions.
func (context *runtimeContext) VerifyContractCode() error {
	if !context.verifyCode {
		return nil
	}

	context.verifyCode = false

	err := context.validator.verifyMemoryDeclaration(context.instance)
	if err != nil {
		logRuntime.Trace("verify contract code", "error", err)
		return err
	}

	err = context.validator.verifyFunctions(context.instance)
	if err != nil {
		logRuntime.Trace("verify contract code", "error", err)
		return err
	}

	err = context.checkBackwardCompatibility()
	if err != nil {
		logRuntime.Trace("verify contract code", "error", err)
		return err
	}

	logRuntime.Trace("verified contract code")

	return nil
}

func (context *runtimeContext) checkBackwardCompatibility() error {
	if context.host.IsDCDTFunctionsEnabled() {
		return nil
	}

	if context.instance.IsFunctionImported("transferDCDTExecute") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("transferDCDTNFTExecute") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("transferValueExecute") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTBalance") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTTokenData") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTTokenType") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTTokenNonce") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getCurrentDCDTNFTNonce") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTNFTNameLength") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTNFTAttributeLength") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("getDCDTNFTURILength") {
		return vmhost.ErrContractInvalid
	}
	if context.instance.IsFunctionImported("bigIntGetDCDTExternalBalance") {
		return vmhost.ErrContractInvalid
	}

	return nil
}

// BaseOpsErrorShouldFailExecution returns true
func (context *runtimeContext) BaseOpsErrorShouldFailExecution() bool {
	return true
}

// SyncExecAPIErrorShouldFailExecution returns true
func (context *runtimeContext) SyncExecAPIErrorShouldFailExecution() bool {
	return true
}

// BigIntAPIErrorShouldFailExecution returns true
func (context *runtimeContext) BigIntAPIErrorShouldFailExecution() bool {
	return true
}

// CryptoAPIErrorShouldFailExecution returns true
func (context *runtimeContext) CryptoAPIErrorShouldFailExecution() bool {
	return true
}

// GetPointsUsed returns the gas points used by the current wasmer instance.
func (context *runtimeContext) GetPointsUsed() uint64 {
	if context.instance == nil {
		return 0
	}
	return context.instance.GetPointsUsed()
}

// SetPointsUsed sets the given gas points as the gas points used by the current wasmer instance.
func (context *runtimeContext) SetPointsUsed(gasPoints uint64) {
	if gasPoints > builtinMath.MaxInt64 {
		gasPoints = builtinMath.MaxInt64
	}
	context.instance.SetPointsUsed(gasPoints)
}

// ReadOnly returns true if the current context is readOnly
func (context *runtimeContext) ReadOnly() bool {
	return context.readOnly
}

// SetReadOnly sets the readOnly field of the context to the given value.
func (context *runtimeContext) SetReadOnly(readOnly bool) {
	context.readOnly = readOnly
}

// GetInstance returns the current wasmer instance
func (context *runtimeContext) GetInstance() wasmer.InstanceHandler {
	return context.instance
}

// GetInstanceExports returns the current wasmer instance exports.
func (context *runtimeContext) GetInstanceExports() wasmer.ExportsMap {
	return context.instance.GetExports()
}

// CleanWasmerInstance cleans the current wasmer instance.
func (context *runtimeContext) CleanWasmerInstance() {
	if context.instance == nil || context.IsWarmInstance() {
		return
	}

	context.instance.Clean()
	context.instance = nil

	logRuntime.Trace("instance cleaned")
}

// IsContractOnTheStack iterates over the state stack to find whether the
// provided SC address is already in execution, below the current instance.
func (context *runtimeContext) IsContractOnTheStack(address []byte) bool {
	for _, state := range context.stateStack {
		if bytes.Equal(address, state.scAddress) {
			return true
		}
	}
	return false
}

// GetFunctionToCall returns the function to call from the wasmer instance exports.
func (context *runtimeContext) GetFunctionToCall() (wasmer.ExportedFunctionCallback, error) {
	exports := context.instance.GetExports()
	logRuntime.Trace("get function to call", "function", context.callFunction)
	if function, ok := exports[context.callFunction]; ok {
		return function, nil
	}

	if context.callFunction == vmhost.CallbackFunctionName {
		// TODO rewrite this condition, until the AsyncContext is merged
		logRuntime.Error("get function to call", "error", vmhost.ErrNilCallbackFunction)
		return nil, vmhost.ErrNilCallbackFunction
	}

	return nil, vmhost.ErrFuncNotFound
}

// GetInitFunction returns the init function from the current wasmer instance exports.
func (context *runtimeContext) GetInitFunction() wasmer.ExportedFunctionCallback {
	exports := context.instance.GetExports()
	if init, ok := exports[vmhost.InitFunctionName]; ok {
		return init
	}

	return nil
}

// ExecuteAsyncCall locks the necessary gas and sets the async call info and a runtime breakpoint value.
func (context *runtimeContext) ExecuteAsyncCall(address []byte, data []byte, value []byte) error {
	metering := context.host.Metering()
	err := metering.UseGasForAsyncStep()
	if err != nil {
		return err
	}

	gasToLock := uint64(0)
	shouldLockGas := context.HasCallbackMethod() || !context.host.IsDynamicGasLockingEnabled()
	if shouldLockGas {
		gasToLock = metering.ComputeGasLockedForAsync()
		err = metering.UseGasBounded(gasToLock)
		if err != nil {
			return err
		}
	}

	context.SetAsyncCallInfo(&vmhost.AsyncCallInfo{
		Destination: address,
		Data:        data,
		GasLimit:    metering.GasLeft(),
		GasLocked:   gasToLock,
		ValueBytes:  value,
	})
	context.SetRuntimeBreakpointValue(vmhost.BreakpointAsyncCall)

	logRuntime.Trace("prepare async call",
		"caller", context.GetSCAddress(),
		"dest", address,
		"value", big.NewInt(0).SetBytes(value),
		"data", data)
	return nil
}

// SetAsyncCallInfo sets the given data as the async call info for the current context.
func (context *runtimeContext) SetAsyncCallInfo(asyncCallInfo *vmhost.AsyncCallInfo) {
	context.asyncCallInfo = asyncCallInfo
}

// AddAsyncContextCall adds the given async call to the asyncContextMap at the given identifier.
func (context *runtimeContext) AddAsyncContextCall(contextIdentifier []byte, asyncCall *vmhost.AsyncGeneratedCall) error {
	_, ok := context.asyncContextInfo.AsyncContextMap[string(contextIdentifier)]
	currentContextMap := context.asyncContextInfo.AsyncContextMap
	if !ok {
		currentContextMap[string(contextIdentifier)] = &vmhost.AsyncContext{
			AsyncCalls: make([]*vmhost.AsyncGeneratedCall, 0),
		}
	}

	currentContextMap[string(contextIdentifier)].AsyncCalls =
		append(currentContextMap[string(contextIdentifier)].AsyncCalls, asyncCall)

	return nil
}

// GetAsyncContextInfo returns the async context info for the current context.
func (context *runtimeContext) GetAsyncContextInfo() *vmhost.AsyncContextInfo {
	return context.asyncContextInfo
}

// GetAsyncContext returns the async context mapped to the given context identifier.
func (context *runtimeContext) GetAsyncContext(contextIdentifier []byte) (*vmhost.AsyncContext, error) {
	asyncContext, ok := context.asyncContextInfo.AsyncContextMap[string(contextIdentifier)]
	if !ok {
		return nil, vmhost.ErrAsyncContextDoesNotExist
	}

	return asyncContext, nil
}

// GetAsyncCallInfo returns the async call info for the current context.
func (context *runtimeContext) GetAsyncCallInfo() *vmhost.AsyncCallInfo {
	return context.asyncCallInfo
}

// HasCallbackMethod returns true if the current wasmer instance exports has a callback method.
func (context *runtimeContext) HasCallbackMethod() bool {
	_, ok := context.instance.GetExports()[vmhost.CallbackFunctionName]
	return ok
}

// IsFunctionImported returns true if the WASM module imports the specified function.
func (context *runtimeContext) IsFunctionImported(name string) bool {
	return context.instance.IsFunctionImported(name)
}

// MemLoad returns the contents from the given offset of the WASM memory.
func (context *runtimeContext) MemLoad(offset int32, length int32) ([]byte, error) {
	if length == 0 {
		return []byte{}, nil
	}

	memory := context.instance.GetInstanceCtxMemory()
	memoryView := memory.Data()
	memoryLength := memory.Length()
	requestedEnd := math.AddInt32(offset, length)

	isOffsetTooSmall := offset < 0
	isOffsetTooLarge := uint32(offset) > memoryLength
	isRequestedEndTooLarge := uint32(requestedEnd) > memoryLength
	isLengthNegative := length < 0

	if isOffsetTooSmall || isOffsetTooLarge {
		return nil, fmt.Errorf("mem load: %w", vmhost.ErrBadBounds)
	}
	if isLengthNegative {
		return nil, fmt.Errorf("mem load: %w", vmhost.ErrNegativeLength)
	}

	result := make([]byte, length)
	if isRequestedEndTooLarge {
		copy(result, memoryView[offset:])
	} else {
		copy(result, memoryView[offset:requestedEnd])
	}

	return result, nil
}

// MemLoadMultiple returns multiple byte slices loaded from the WASM memory, starting at the given offset and having the provided lengths.
func (context *runtimeContext) MemLoadMultiple(offset int32, lengths []int32) ([][]byte, error) {
	if len(lengths) == 0 {
		return [][]byte{}, nil
	}

	results := make([][]byte, len(lengths))

	for i, length := range lengths {
		result, err := context.MemLoad(offset, length)
		if err != nil {
			return nil, err
		}

		results[i] = result
		offset += length
	}

	return results, nil
}

// MemStore stores the given data in the WASM memory at the given offset.
func (context *runtimeContext) MemStore(offset int32, data []byte) error {
	dataLength := int32(len(data))
	if dataLength == 0 {
		return nil
	}

	memory := context.instance.GetInstanceCtxMemory()
	memoryView := memory.Data()
	memoryLength := memory.Length()
	requestedEnd := math.AddInt32(offset, dataLength)

	isOffsetTooSmall := offset < 0
	isNewPageNecessary := uint32(requestedEnd) > memoryLength

	if isOffsetTooSmall {
		return vmhost.ErrBadLowerBounds
	}
	if isNewPageNecessary {
		err := memory.Grow(1)
		if err != nil {
			return err
		}

		memoryView = memory.Data()
		memoryLength = memory.Length()
	}

	isRequestedEndTooLarge := uint32(requestedEnd) > memoryLength
	if isRequestedEndTooLarge {
		return vmhost.ErrBadUpperBounds
	}

	copy(memoryView[offset:requestedEnd], data)
	return nil
}

// AddError adds an error to the global error list on runtime context
func (context *runtimeContext) AddError(err error, otherInfo ...string) {
	if err == nil {
		return
	}
	if context.errors == nil {
		context.errors = vmhost.WrapError(err, otherInfo...)
		return
	}
	context.errors = context.errors.WrapWithError(err, otherInfo...)
}

func (context *runtimeContext) GetAllErrors() error {
	return context.errors
}

// SetWarmInstance overwrites the warm Wasmer instance with the provided one.
// TODO remove after implementing proper mocking of Wasmer instances; this is
// used for tests only
// func (context *runtimeContext) SetWarmInstance(address []byte, instanc e
