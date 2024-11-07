package worldmock

import (
	"errors"
	"fmt"

	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/config"
)

// NewAddressMock allows tests to specify what new addresses to generate
type NewAddressMock struct {
	CreatorAddress []byte
	CreatorNonce   uint64
	NewAddress     []byte
}

// BlockInfo contains metadata about a mocked block
type BlockInfo struct {
	BlockTimestamp uint64
	BlockNonce     uint64
	BlockRound     uint64
	BlockEpoch     uint32
	RandomSeed     *[48]byte
}

// MockWorld provides a mock representation of the blockchain to be used in VM tests.
type MockWorld struct {
	SelfShardID                uint32
	AcctMap                    AccountMap
	AccountsAdapter            vmcommon.AccountsAdapter
	PreviousBlockInfo          *BlockInfo
	CurrentBlockInfo           *BlockInfo
	Blockhashes                [][]byte
	NewAddressMocks            []*NewAddressMock
	StateRootHash              []byte
	Err                        error
	LastCreatedContractAddress []byte
	CompiledCode               map[string][]byte
	BuiltinFuncs               *BuiltinFunctionsWrapper
	GuardedAccountHandler      vmcommon.GuardedAccountHandler
}

// NewMockWorld creates a new MockWorld instance
func NewMockWorld() *MockWorld {
	accountMap := NewAccountMap()
	world := &MockWorld{
		SelfShardID:           0,
		AcctMap:               accountMap,
		AccountsAdapter:       nil,
		PreviousBlockInfo:     nil,
		CurrentBlockInfo:      &BlockInfo{},
		Blockhashes:           nil,
		NewAddressMocks:       nil,
		CompiledCode:          make(map[string][]byte),
		BuiltinFuncs:          nil,
		GuardedAccountHandler: nil,
	}
	world.AccountsAdapter = NewMockAccountsAdapter(world)
	world.GuardedAccountHandler = NewMockGuardedAccountHandler()

	return world
}

// InitBuiltinFunctions initializes the inner BuiltinFunctionsWrapper, required
// for calling builtin functions.
func (b *MockWorld) InitBuiltinFunctions(gasMap config.GasScheduleMap) error {
	wrapper, err := NewBuiltinFunctionsWrapper(b, gasMap)
	if err != nil {
		return err
	}

	b.BuiltinFuncs = wrapper
	return nil
}

// Clear resets all mock data between tests.
func (b *MockWorld) Clear() {
	b.AcctMap = NewAccountMap()
	b.AccountsAdapter = NewMockAccountsAdapter(b)
	b.PreviousBlockInfo = nil
	b.CurrentBlockInfo = nil
	b.Blockhashes = nil
	b.NewAddressMocks = nil
	b.CompiledCode = make(map[string][]byte)
}

// SetCurrentBlockHash -
func (b *MockWorld) SetCurrentBlockHash(blockHash []byte) {
	if b.CurrentBlockInfo == nil {
		b.CurrentBlockInfo = &BlockInfo{}
	}
	b.Blockhashes = [][]byte{blockHash}
}

// NumberOfShards -
func (b *MockWorld) NumberOfShards() uint32 {
	maxShardID := uint32(0)
	for _, account := range b.AcctMap {
		if account.ShardID > maxShardID {
			maxShardID = account.ShardID
		}
	}

	return maxShardID + 1
}

// ComputeId -
func (b *MockWorld) ComputeId(address []byte) uint32 {
	return b.AcctMap.GetAccount(address).ShardID
}

// SelfId -
func (b *MockWorld) SelfId() uint32 {
	return b.SelfShardID
}

// SameShard -
func (b *MockWorld) SameShard(firstAddress []byte, secondAddress []byte) bool {
	firstAccount := b.AcctMap.GetAccount(firstAddress)
	secondAccount := b.AcctMap.GetAccount(secondAddress)
	return firstAccount.ShardID == secondAccount.ShardID
}

// CommunicationIdentifier -
func (b *MockWorld) CommunicationIdentifier(destShardID uint32) string {
	return fmt.Sprintf("commID-dest-%d", destShardID)
}

// GetSnapshot -
func (b *MockWorld) GetSnapshot() int {
	return b.AccountsAdapter.JournalLen()
}

// RevertToSnapshot -
func (b *MockWorld) RevertToSnapshot(snapshot int) error {
	return b.AccountsAdapter.RevertToSnapshot(snapshot)
}

// ExecuteSmartContractCallOnOtherVM -
func (b *MockWorld) ExecuteSmartContractCallOnOtherVM(_ *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	return nil, errors.New("not implemented")
}
