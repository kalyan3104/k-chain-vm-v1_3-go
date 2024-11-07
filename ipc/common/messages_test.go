package common

import (
	"reflect"
	"testing"

	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/ipc/marshaling"
	"github.com/stretchr/testify/require"
)

func TestMessageContractResponse_IsConsistentlySerializable(t *testing.T) {
	vmOutput := &vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	vmOutput.OutputAccounts["alice"] = &vmcommon.OutputAccount{StorageUpdates: make(map[string]*vmcommon.StorageUpdate)}
	vmOutput.OutputAccounts["alice"].StorageUpdates["foo"] = &vmcommon.StorageUpdate{}
	vmOutput.OutputAccounts["alice"].StorageUpdates["bar"] = &vmcommon.StorageUpdate{}
	message := NewMessageContractResponse(vmOutput, nil)
	requireSerializationConsistency(t, message, &MessageContractResponse{})

	// Non text as output account keys
	vmOutput = &vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	vmOutput.OutputAccounts[string([]byte{0})] = &vmcommon.OutputAccount{StorageUpdates: make(map[string]*vmcommon.StorageUpdate)}
	vmOutput.OutputAccounts[string([]byte{0})].StorageUpdates["foo"] = &vmcommon.StorageUpdate{}
	vmOutput.OutputAccounts[string([]byte{0})].StorageUpdates["bar"] = &vmcommon.StorageUpdate{}
	message = NewMessageContractResponse(vmOutput, nil)
	requireSerializationConsistency(t, message, &MessageContractResponse{})

	// Non UTF-8 as output account keys
	vmOutput = &vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	vmOutput.OutputAccounts[string([]byte{128})] = &vmcommon.OutputAccount{StorageUpdates: make(map[string]*vmcommon.StorageUpdate)}
	vmOutput.OutputAccounts[string([]byte{128})].StorageUpdates["foo"] = &vmcommon.StorageUpdate{}
	vmOutput.OutputAccounts[string([]byte{128})].StorageUpdates["bar"] = &vmcommon.StorageUpdate{}
	message = NewMessageContractResponse(vmOutput, nil)
	requireSerializationConsistency(t, message, &MessageContractResponse{})

	// Non UTF-8 as storage update keys
	vmOutput = &vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	vmOutput.OutputAccounts["alice"] = &vmcommon.OutputAccount{StorageUpdates: make(map[string]*vmcommon.StorageUpdate)}
	vmOutput.OutputAccounts["alice"].StorageUpdates[string([]byte{128})] = &vmcommon.StorageUpdate{}
	vmOutput.OutputAccounts["alice"].StorageUpdates[string([]byte{129})] = &vmcommon.StorageUpdate{}
	message = NewMessageContractResponse(vmOutput, nil)
	requireSerializationConsistency(t, message, &MessageContractResponse{})
}

func TestMessageContractResponse_CanWrapNilVMOutput(t *testing.T) {
	message := NewMessageContractResponse(nil, nil)
	expectedEmptyVMOutput := vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	actualVMOutput := *message.SerializableVMOutput.ConvertToVMOutput()

	require.True(t, reflect.DeepEqual(expectedEmptyVMOutput, actualVMOutput))
	requireSerializationConsistency(t, message, &MessageContractResponse{})
}

func TestMessageBlockchainProcessBuiltinFunctionResponse_IsConsistentlySerializable(t *testing.T) {
	vmOutput := &vmcommon.VMOutput{OutputAccounts: make(map[string]*vmcommon.OutputAccount)}
	vmOutput.OutputAccounts["alice"] = &vmcommon.OutputAccount{Address: []byte("alice")}
	// Non UTF-8 as output account keys
	vmOutput.OutputAccounts[string([]byte{0, 129})] = &vmcommon.OutputAccount{Address: []byte{0, 129}}
	vmOutput.OutputAccounts[string([]byte{0, 128})] = &vmcommon.OutputAccount{Address: []byte{0, 128}}
	message := NewMessageBlockchainProcessBuiltInFunctionResponse(vmOutput, nil)
	requireSerializationConsistency(t, message, &MessageBlockchainProcessBuiltInFunctionResponse{})
}

func TestMessageBlockchainGetAllStateResponse_IsConsistentlySerializable(t *testing.T) {
	allState := make(map[string][]byte)
	allState["foo"] = []byte{0}
	allState[string([]byte{0})] = []byte{0}
	allState[string([]byte{128})] = []byte{0}
	message := NewMessageBlockchainGetAllStateResponse(allState, nil)
	requireSerializationConsistency(t, message, &MessageBlockchainGetAllStateResponse{})
}

func requireSerializationConsistency(t *testing.T, message interface{}, intoMessage interface{}) {
	marshalizer := marshaling.CreateMarshalizer(marshaling.JSON)

	serialized, err := marshalizer.Marshal(message)
	require.Nil(t, err)
	err = marshalizer.Unmarshal(intoMessage, serialized)
	require.Nil(t, err)

	areEqual := reflect.DeepEqual(message, intoMessage)
	if !areEqual {
		require.FailNow(t, "Serialization is not consistent.")
	}
}
