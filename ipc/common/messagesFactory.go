package common

// CreateMessage creates a message given its kind
func CreateMessage(kind MessageKind) MessageHandler {
	kindIndex := uint32(kind)
	length := uint32(len(messageCreators))
	if kindIndex < length {
		message := messageCreators[kindIndex]()
		message.SetKind(kind)
		return message
	}

	return createUndefinedMessage()
}

type messageCreator func() MessageHandler

var messageCreators = make([]messageCreator, LastKind)

func init() {
	for i := 0; i < len(messageCreators); i++ {
		messageCreators[i] = createUndefinedMessage
	}

	messageCreators[Initialize] = createMessageInitialize
	messageCreators[Stop] = createMessageStop
	messageCreators[ContractDeployRequest] = createMessageContractDeployRequest
	messageCreators[ContractCallRequest] = createMessageContractCallRequest
	messageCreators[ContractResponse] = createMessageContractResponse
	messageCreators[DiagnoseWaitRequest] = createMessageDiagnoseWaitRequest
	messageCreators[DiagnoseWaitResponse] = createMessageDiagnoseWaitResponse
	messageCreators[VersionRequest] = createMessageVersionRequest
	messageCreators[VersionResponse] = createMessageVersionResponse

	messageCreators[BlockchainNewAddressRequest] = createMessageBlockchainNewAddressRequest
	messageCreators[BlockchainNewAddressResponse] = createMessageBlockchainNewAddressResponse
	messageCreators[BlockchainGetStorageDataRequest] = createMessageBlockchainGetStorageDataRequest
	messageCreators[BlockchainGetStorageDataResponse] = createMessageBlockchainGetStorageDataResponse
	messageCreators[BlockchainGetBlockhashRequest] = createMessageBlockchainGetBlockhashRequest
	messageCreators[BlockchainGetBlockhashResponse] = createMessageBlockchainGetBlockhashResponse
	messageCreators[BlockchainLastNonceRequest] = createMessageBlockchainLastNonceRequest
	messageCreators[BlockchainLastNonceResponse] = createMessageBlockchainLastNonceResponse
	messageCreators[BlockchainLastRoundRequest] = createMessageBlockchainLastRoundRequest
	messageCreators[BlockchainLastRoundResponse] = createMessageBlockchainLastRoundResponse
	messageCreators[BlockchainLastTimeStampRequest] = createMessageBlockchainLastTimeStampRequest
	messageCreators[BlockchainLastTimeStampResponse] = createMessageBlockchainLastTimeStampResponse
	messageCreators[BlockchainLastRandomSeedRequest] = createMessageBlockchainLastRandomSeedRequest
	messageCreators[BlockchainLastRandomSeedResponse] = createMessageBlockchainLastRandomSeedResponse
	messageCreators[BlockchainLastEpochRequest] = createMessageBlockchainLastEpochRequest
	messageCreators[BlockchainLastEpochResponse] = createMessageBlockchainLastEpochResponse
	messageCreators[BlockchainGetStateRootHashRequest] = createMessageBlockchainGetStateRootHashRequest
	messageCreators[BlockchainGetStateRootHashResponse] = createMessageBlockchainGetStateRootHashResponse
	messageCreators[BlockchainCurrentNonceRequest] = createMessageBlockchainCurrentNonceRequest
	messageCreators[BlockchainCurrentNonceResponse] = createMessageBlockchainCurrentNonceResponse
	messageCreators[BlockchainCurrentRoundRequest] = createMessageBlockchainCurrentRoundRequest
	messageCreators[BlockchainCurrentRoundResponse] = createMessageBlockchainCurrentRoundResponse
	messageCreators[BlockchainCurrentTimeStampRequest] = createMessageBlockchainCurrentTimeStampRequest
	messageCreators[BlockchainCurrentTimeStampResponse] = createMessageBlockchainCurrentTimeStampResponse
	messageCreators[BlockchainCurrentRandomSeedRequest] = createMessageBlockchainCurrentRandomSeedRequest
	messageCreators[BlockchainCurrentRandomSeedResponse] = createMessageBlockchainCurrentRandomSeedResponse
	messageCreators[BlockchainCurrentEpochRequest] = createMessageBlockchainCurrentEpochRequest
	messageCreators[BlockchainCurrentEpochResponse] = createMessageBlockchainCurrentEpochResponse
	messageCreators[BlockchainProcessBuiltInFunctionRequest] = createMessageBlockchainProcessBuiltInFunctionRequest
	messageCreators[BlockchainProcessBuiltInFunctionResponse] = createMessageBlockchainProcessBuiltInFunctionResponse
	messageCreators[BlockchainGetBuiltinFunctionNamesRequest] = createMessageBlockchainGetBuiltinFunctionNamesRequest
	messageCreators[BlockchainGetBuiltinFunctionNamesResponse] = createMessageBlockchainGetBuiltinFunctionNamesResponse
	messageCreators[BlockchainGetAllStateRequest] = createMessageBlockchainGetAllStateRequest
	messageCreators[BlockchainGetAllStateResponse] = createMessageBlockchainGetAllStateResponse
	messageCreators[BlockchainGetUserAccountRequest] = createMessageBlockchainGetUserAccountRequest
	messageCreators[BlockchainGetUserAccountResponse] = createMessageBlockchainGetUserAccountResponse
	messageCreators[BlockchainGetCodeRequest] = createMessageBlockchainGetCodeRequest
	messageCreators[BlockchainGetCodeResponse] = createMessageBlockchainGetCodeResponse
	messageCreators[BlockchainGetShardOfAddressRequest] = createMessageBlockchainGetShardOfAddressRequest
	messageCreators[BlockchainGetShardOfAddressResponse] = createMessageBlockchainGetShardOfAddressResponse
	messageCreators[BlockchainIsSmartContractRequest] = createMessageBlockchainIsSmartContractRequest
	messageCreators[BlockchainIsSmartContractResponse] = createMessageBlockchainIsSmartContractResponse
	messageCreators[BlockchainIsPayableRequest] = createMessageBlockchainIsPayableRequest
	messageCreators[BlockchainIsPayableResponse] = createMessageBlockchainIsPayableResponse
	messageCreators[BlockchainSaveCompiledCodeRequest] = createMessageBlockchainSaveCompiledCodeRequest
	messageCreators[BlockchainSaveCompiledCodeResponse] = createMessageBlockchainSaveCompiledCodeResponse
	messageCreators[BlockchainGetCompiledCodeRequest] = createMessageBlockchainGetCompiledCodeRequest
	messageCreators[BlockchainGetCompiledCodeResponse] = createMessageBlockchainGetCompiledCodeResponse
	messageCreators[BlockchainClearCompiledCodesRequest] = createMessageBlockchainClearCompiledCodesRequest
	messageCreators[BlockchainClearCompiledCodesResponse] = createMessageBlockchainClearCompiledCodesResponse
	messageCreators[BlockchainGetDCDTTokenRequest] = createMessageBlockchainGetDCDTTokenRequest
	messageCreators[BlockchainGetDCDTTokenResponse] = createMessageBlockchainGetDCDTTokenResponse
	messageCreators[BlockchainIsInterfaceNilRequest] = createMessageBlockchainIsInterfaceNilRequest
	messageCreators[BlockchainIsInterfaceNilResponse] = createMessageBlockchainIsInterfaceNilResponse
	messageCreators[BlockchainGetSnapshotRequest] = createMessageBlockchainGetSnapshotRequest
	messageCreators[BlockchainGetSnapshotResponse] = createMessageBlockchainGetSnapshotResponse
	messageCreators[BlockchainRevertToSnapshotRequest] = createMessageBlockchainRevertToSnapshotRequest
	messageCreators[BlockchainRevertToSnapshotResponse] = createMessageBlockchainRevertToSnapshotResponse

}

func createMessageInitialize() MessageHandler {
	return &MessageInitialize{}
}

func createMessageStop() MessageHandler {
	return &MessageStop{}
}

func createMessageContractDeployRequest() MessageHandler {
	return &MessageContractDeployRequest{}
}

func createMessageContractCallRequest() MessageHandler {
	return &MessageContractCallRequest{}
}

func createMessageContractResponse() MessageHandler {
	return &MessageContractResponse{}
}

func createMessageDiagnoseWaitRequest() MessageHandler {
	return &MessageDiagnoseWaitRequest{}
}

func createMessageDiagnoseWaitResponse() MessageHandler {
	return &MessageDiagnoseWaitResponse{}
}

func createMessageVersionRequest() MessageHandler {
	return &MessageVersionRequest{}
}

func createMessageVersionResponse() MessageHandler {
	return &MessageVersionResponse{}
}

func createUndefinedMessage() MessageHandler {
	return NewUndefinedMessage()
}

func createMessageBlockchainNewAddressRequest() MessageHandler {
	return &MessageBlockchainNewAddressRequest{}
}

func createMessageBlockchainNewAddressResponse() MessageHandler {
	return &MessageBlockchainNewAddressResponse{}
}

func createMessageBlockchainGetStorageDataRequest() MessageHandler {
	return &MessageBlockchainGetStorageDataRequest{}
}

func createMessageBlockchainGetStorageDataResponse() MessageHandler {
	return &MessageBlockchainGetStorageDataResponse{}
}

func createMessageBlockchainGetBlockhashRequest() MessageHandler {
	return &MessageBlockchainGetBlockhashRequest{}
}

func createMessageBlockchainGetBlockhashResponse() MessageHandler {
	return &MessageBlockchainGetBlockhashResponse{}
}

func createMessageBlockchainLastNonceRequest() MessageHandler {
	return &MessageBlockchainLastNonceRequest{}
}

func createMessageBlockchainLastNonceResponse() MessageHandler {
	return &MessageBlockchainLastNonceResponse{}
}

func createMessageBlockchainLastRoundRequest() MessageHandler {
	return &MessageBlockchainLastRoundRequest{}
}

func createMessageBlockchainLastRoundResponse() MessageHandler {
	return &MessageBlockchainLastRoundResponse{}
}

func createMessageBlockchainLastTimeStampRequest() MessageHandler {
	return &MessageBlockchainLastTimeStampRequest{}
}

func createMessageBlockchainLastTimeStampResponse() MessageHandler {
	return &MessageBlockchainLastTimeStampResponse{}
}

func createMessageBlockchainLastRandomSeedRequest() MessageHandler {
	return &MessageBlockchainLastRandomSeedRequest{}
}

func createMessageBlockchainLastRandomSeedResponse() MessageHandler {
	return &MessageBlockchainLastRandomSeedResponse{}
}

func createMessageBlockchainLastEpochRequest() MessageHandler {
	return &MessageBlockchainLastEpochRequest{}
}

func createMessageBlockchainLastEpochResponse() MessageHandler {
	return &MessageBlockchainLastEpochResponse{}
}

func createMessageBlockchainGetStateRootHashRequest() MessageHandler {
	return &MessageBlockchainGetStateRootHashRequest{}
}

func createMessageBlockchainGetStateRootHashResponse() MessageHandler {
	return &MessageBlockchainGetStateRootHashResponse{}
}

func createMessageBlockchainCurrentNonceRequest() MessageHandler {
	return &MessageBlockchainCurrentNonceRequest{}
}

func createMessageBlockchainCurrentNonceResponse() MessageHandler {
	return &MessageBlockchainCurrentNonceResponse{}
}

func createMessageBlockchainCurrentRoundRequest() MessageHandler {
	return &MessageBlockchainCurrentRoundRequest{}
}

func createMessageBlockchainCurrentRoundResponse() MessageHandler {
	return &MessageBlockchainCurrentRoundResponse{}
}

func createMessageBlockchainCurrentTimeStampRequest() MessageHandler {
	return &MessageBlockchainCurrentTimeStampRequest{}
}

func createMessageBlockchainCurrentTimeStampResponse() MessageHandler {
	return &MessageBlockchainCurrentTimeStampResponse{}
}

func createMessageBlockchainCurrentRandomSeedRequest() MessageHandler {
	return &MessageBlockchainCurrentRandomSeedRequest{}
}

func createMessageBlockchainCurrentRandomSeedResponse() MessageHandler {
	return &MessageBlockchainCurrentRandomSeedResponse{}
}

func createMessageBlockchainCurrentEpochRequest() MessageHandler {
	return &MessageBlockchainCurrentEpochRequest{}
}

func createMessageBlockchainCurrentEpochResponse() MessageHandler {
	return &MessageBlockchainCurrentEpochResponse{}
}

func createMessageBlockchainProcessBuiltInFunctionRequest() MessageHandler {
	return &MessageBlockchainProcessBuiltInFunctionRequest{}
}

func createMessageBlockchainProcessBuiltInFunctionResponse() MessageHandler {
	return &MessageBlockchainProcessBuiltInFunctionResponse{}
}

func createMessageBlockchainGetBuiltinFunctionNamesRequest() MessageHandler {
	return &MessageBlockchainGetBuiltinFunctionNamesRequest{}
}

func createMessageBlockchainGetBuiltinFunctionNamesResponse() MessageHandler {
	return &MessageBlockchainGetBuiltinFunctionNamesResponse{}
}

func createMessageBlockchainGetAllStateRequest() MessageHandler {
	return &MessageBlockchainGetAllStateRequest{}
}

func createMessageBlockchainGetAllStateResponse() MessageHandler {
	return &MessageBlockchainGetAllStateResponse{}
}

func createMessageBlockchainGetUserAccountRequest() MessageHandler {
	return &MessageBlockchainGetUserAccountRequest{}
}

func createMessageBlockchainGetUserAccountResponse() MessageHandler {
	return &MessageBlockchainGetUserAccountResponse{}
}

func createMessageBlockchainGetCodeRequest() MessageHandler {
	return &MessageBlockchainGetCodeRequest{}
}

func createMessageBlockchainGetCodeResponse() MessageHandler {
	return &MessageBlockchainGetCodeResponse{}
}

func createMessageBlockchainGetShardOfAddressRequest() MessageHandler {
	return &MessageBlockchainGetShardOfAddressRequest{}
}

func createMessageBlockchainGetShardOfAddressResponse() MessageHandler {
	return &MessageBlockchainGetShardOfAddressResponse{}
}

func createMessageBlockchainIsSmartContractRequest() MessageHandler {
	return &MessageBlockchainIsSmartContractRequest{}
}

func createMessageBlockchainIsSmartContractResponse() MessageHandler {
	return &MessageBlockchainIsSmartContractResponse{}
}

func createMessageBlockchainIsPayableRequest() MessageHandler {
	return &MessageBlockchainIsPayableRequest{}
}

func createMessageBlockchainIsPayableResponse() MessageHandler {
	return &MessageBlockchainIsPayableResponse{}
}

func createMessageBlockchainSaveCompiledCodeRequest() MessageHandler {
	return &MessageBlockchainSaveCompiledCodeRequest{}
}

func createMessageBlockchainSaveCompiledCodeResponse() MessageHandler {
	return &MessageBlockchainSaveCompiledCodeResponse{}
}

func createMessageBlockchainGetCompiledCodeRequest() MessageHandler {
	return &MessageBlockchainGetCompiledCodeRequest{}
}

func createMessageBlockchainGetCompiledCodeResponse() MessageHandler {
	return &MessageBlockchainGetCompiledCodeResponse{}
}

func createMessageBlockchainClearCompiledCodesRequest() MessageHandler {
	return &MessageBlockchainClearCompiledCodesRequest{}
}

func createMessageBlockchainClearCompiledCodesResponse() MessageHandler {
	return &MessageBlockchainClearCompiledCodesResponse{}
}

func createMessageBlockchainGetDCDTTokenRequest() MessageHandler {
	return &MessageBlockchainGetDCDTTokenRequest{}
}

func createMessageBlockchainGetDCDTTokenResponse() MessageHandler {
	return &MessageBlockchainGetDCDTTokenResponse{}
}

func createMessageBlockchainIsInterfaceNilRequest() MessageHandler {
	return &MessageBlockchainIsInterfaceNilRequest{}
}

func createMessageBlockchainIsInterfaceNilResponse() MessageHandler {
	return &MessageBlockchainIsInterfaceNilResponse{}
}

func createMessageBlockchainGetSnapshotRequest() MessageHandler {
	return &MessageBlockchainGetSnapshotRequest{}
}

func createMessageBlockchainGetSnapshotResponse() MessageHandler {
	return &MessageBlockchainGetSnapshotResponse{}
}

func createMessageBlockchainRevertToSnapshotRequest() MessageHandler {
	return &MessageBlockchainRevertToSnapshotRequest{}
}

func createMessageBlockchainRevertToSnapshotResponse() MessageHandler {
	return &MessageBlockchainRevertToSnapshotResponse{}
}
