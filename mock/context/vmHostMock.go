package mock

import (
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/data/vm"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/config"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/crypto"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/wasmer"
)

var _ vmhost.VMHost = (*VMHostMock)(nil)

// VMHostMock is used in tests to check the VMHost interface method calls
type VMHostMock struct {
	BlockChainHook vmcommon.BlockchainHook
	CryptoHook     crypto.VMCrypto

	EthInput []byte

	BlockchainContext vmhost.BlockchainContext
	RuntimeContext    vmhost.RuntimeContext
	OutputContext     vmhost.OutputContext
	MeteringContext   vmhost.MeteringContext
	StorageContext    vmhost.StorageContext
	BigIntContext     vmhost.BigIntContext

	SCAPIMethods  *wasmer.Imports
	IsBuiltinFunc bool
}

// GetVersion mocked method
func (host *VMHostMock) GetVersion() string {
	return "mock"
}

// Crypto mocked method
func (host *VMHostMock) Crypto() crypto.VMCrypto {
	return host.CryptoHook
}

// Blockchain mocked method
func (host *VMHostMock) Blockchain() vmhost.BlockchainContext {
	return host.BlockchainContext
}

// Runtime mocked method
func (host *VMHostMock) Runtime() vmhost.RuntimeContext {
	return host.RuntimeContext
}

// Output mocked method
func (host *VMHostMock) Output() vmhost.OutputContext {
	return host.OutputContext
}

// Metering mocked method
func (host *VMHostMock) Metering() vmhost.MeteringContext {
	return host.MeteringContext
}

// Storage mocked method
func (host *VMHostMock) Storage() vmhost.StorageContext {
	return host.StorageContext
}

// BigInt mocked method
func (host *VMHostMock) BigInt() vmhost.BigIntContext {
	return host.BigIntContext
}

// IsVMV2Enabled mocked method
func (host *VMHostMock) IsVMV2Enabled() bool {
	return true
}

// IsVMV3Enabled mocked method
func (host *VMHostMock) IsVMV3Enabled() bool {
	return true
}

// IsAheadOfTimeCompileEnabled mocked method
func (host *VMHostMock) IsAheadOfTimeCompileEnabled() bool {
	return true
}

// IsDynamicGasLockingEnabled mocked method
func (host *VMHostMock) IsDynamicGasLockingEnabled() bool {
	return true
}

// IsDCDTFunctionsEnabled mocked method
func (host *VMHostMock) IsDCDTFunctionsEnabled() bool {
	return true
}

// AreInSameShard mocked method
func (host *VMHostMock) AreInSameShard(_ []byte, _ []byte) bool {
	return true
}

// ExecuteDCDTTransfer mocked method
func (host *VMHostMock) ExecuteDCDTTransfer(_ []byte, _ []byte, _ []byte, _ uint64, _ *big.Int, _ vm.CallType) (*vmcommon.VMOutput, uint64, error) {
	return nil, 0, nil
}

// CreateNewContract mocked method
func (host *VMHostMock) CreateNewContract(_ *vmcommon.ContractCreateInput) ([]byte, error) {
	return nil, nil
}

// ExecuteOnSameContext mocked method
func (host *VMHostMock) ExecuteOnSameContext(_ *vmcommon.ContractCallInput) (*vmhost.AsyncContextInfo, error) {
	return nil, nil
}

// ExecuteOnDestContext mocked method
func (host *VMHostMock) ExecuteOnDestContext(_ *vmcommon.ContractCallInput) (*vmcommon.VMOutput, *vmhost.AsyncContextInfo, error) {
	return nil, nil, nil
}

// InitState mocked method
func (host *VMHostMock) InitState() {
}

// PushState mocked method
func (host *VMHostMock) PushState() {
}

// PopState mocked method
func (host *VMHostMock) PopState() {
}

// ClearStateStack mocked method
func (host *VMHostMock) ClearStateStack() {
}

// GetAPIMethods mocked method
func (host *VMHostMock) GetAPIMethods() *wasmer.Imports {
	return host.SCAPIMethods
}

// IsBuiltinFunctionName mocked method
func (host *VMHostMock) IsBuiltinFunctionName(_ string) bool {
	return host.IsBuiltinFunc
}

// GetGasScheduleMap mocked method
func (host *VMHostMock) GetGasScheduleMap() config.GasScheduleMap {
	return make(config.GasScheduleMap)
}

// RunSmartContractCall mocked method
func (host *VMHostMock) RunSmartContractCall(_ *vmcommon.ContractCallInput) (vmOutput *vmcommon.VMOutput, err error) {
	return nil, nil
}

// Close -
func (host *VMHostMock) Close() error {
	return nil
}

// RunSmartContractCreate mocked method
func (host *VMHostMock) RunSmartContractCreate(_ *vmcommon.ContractCreateInput) (vmOutput *vmcommon.VMOutput, err error) {
	return nil, nil
}

// GasScheduleChange mocked method
func (host *VMHostMock) GasScheduleChange(_ config.GasScheduleMap) {
}

// SetBuiltInFunctionsContainer mocked method
func (host *VMHostMock) SetBuiltInFunctionsContainer(_ vmcommon.BuiltInFunctionContainer) {
}

// IsInterfaceNil mocked method
func (host *VMHostMock) IsInterfaceNil() bool {
	return false
}

// GetContexts mocked method
func (host *VMHostMock) GetContexts() (
	vmhost.BigIntContext,
	vmhost.BlockchainContext,
	vmhost.MeteringContext,
	vmhost.OutputContext,
	vmhost.RuntimeContext,
	vmhost.StorageContext,
) {
	return host.BigIntContext, host.BlockchainContext, host.MeteringContext, host.OutputContext, host.RuntimeContext, host.StorageContext
}

// SetRuntimeContext mocked method
func (host *VMHostMock) SetRuntimeContext(runtime vmhost.RuntimeContext) {
	host.RuntimeContext = runtime
}
