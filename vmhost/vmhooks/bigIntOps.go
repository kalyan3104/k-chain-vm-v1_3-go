package vmhooks

// // Declare the function signatures (see [cgo](https://golang.org/cmd/cgo/)).
//
// #include <stdlib.h>
// typedef unsigned char uint8_t;
// typedef int int32_t;
//
// extern int32_t		v1_3_bigIntNew(void* context, long long smallValue);
//
// extern int32_t		v1_3_bigIntUnsignedByteLength(void* context, int32_t reference);
// extern int32_t		v1_3_bigIntSignedByteLength(void* context, int32_t reference);
// extern int32_t		v1_3_bigIntGetUnsignedBytes(void* context, int32_t reference, int32_t byteOffset);
// extern int32_t		v1_3_bigIntGetSignedBytes(void* context, int32_t reference, int32_t byteOffset);
// extern void			v1_3_bigIntSetUnsignedBytes(void* context, int32_t destination, int32_t byteOffset, int32_t byteLength);
// extern void			v1_3_bigIntSetSignedBytes(void* context, int32_t destination, int32_t byteOffset, int32_t byteLength);
//
// extern int32_t		v1_3_bigIntIsInt64(void* context, int32_t reference);
// extern long long	v1_3_bigIntGetInt64(void* context, int32_t reference);
// extern void			v1_3_bigIntSetInt64(void* context, int32_t destination, long long value);
//
// extern void			v1_3_bigIntAdd(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntSub(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntMul(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntTDiv(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntTMod(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntEDiv(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntEMod(void* context, int32_t destination, int32_t op1, int32_t op2);
//
// extern void			v1_3_bigIntAbs(void* context, int32_t destination, int32_t op);
// extern void			v1_3_bigIntNeg(void* context, int32_t destination, int32_t op);
// extern int32_t		v1_3_bigIntSign(void* context, int32_t op);
// extern int32_t		v1_3_bigIntCmp(void* context, int32_t op1, int32_t op2);
//
// extern void			v1_3_bigIntNot(void* context, int32_t destination, int32_t op);
// extern void			v1_3_bigIntAnd(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntOr(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntXor(void* context, int32_t destination, int32_t op1, int32_t op2);
// extern void			v1_3_bigIntShr(void* context, int32_t destination, int32_t op, int32_t bits);
// extern void			v1_3_bigIntShl(void* context, int32_t destination, int32_t op, int32_t bits);
//
// extern void			v1_3_bigIntFinishUnsigned(void* context, int32_t reference);
// extern void			v1_3_bigIntFinishSigned(void* context, int32_t reference);
// extern int32_t		v1_3_bigIntStorageStoreUnsigned(void *context, int32_t keyOffset, int32_t keyLength, int32_t source);
// extern int32_t		v1_3_bigIntStorageLoadUnsigned(void *context, int32_t keyOffset, int32_t keyLength, int32_t destination);
// extern void			v1_3_bigIntGetUnsignedArgument(void *context, int32_t id, int32_t destination);
// extern void			v1_3_bigIntGetSignedArgument(void *context, int32_t id, int32_t destination);
// extern void			v1_3_bigIntGetCallValue(void *context, int32_t destination);
// extern void			v1_3_bigIntGetDCDTCallValue(void *context, int32_t destination);
// extern void			v1_3_bigIntGetDCDTExternalBalance(void *context, int32_t addressOffset, int32_t tokenIDOffset, int32_t tokenIDLen, long long nonce, int32_t result);
// extern void			v1_3_bigIntGetExternalBalance(void *context, int32_t addressOffset, int32_t result);
import "C"

import (
	"math/big"
	"unsafe"

	"github.com/kalyan3104/k-chain-vm-v1_3-go/math"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/wasmer"
	twos "github.com/kalyan3104/k-components-big-int/twos-complement"
)

// BigIntImports creates a new wasmer.Imports populated with the BigInt API methods
func BigIntImports(imports *wasmer.Imports) (*wasmer.Imports, error) {
	imports = imports.Namespace("env")

	imports, err := imports.Append("bigIntNew", v1_3_bigIntNew, C.v1_3_bigIntNew)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntUnsignedByteLength", v1_3_bigIntUnsignedByteLength, C.v1_3_bigIntUnsignedByteLength)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSignedByteLength", v1_3_bigIntSignedByteLength, C.v1_3_bigIntSignedByteLength)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetUnsignedBytes", v1_3_bigIntGetUnsignedBytes, C.v1_3_bigIntGetUnsignedBytes)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetSignedBytes", v1_3_bigIntGetSignedBytes, C.v1_3_bigIntGetSignedBytes)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSetUnsignedBytes", v1_3_bigIntSetUnsignedBytes, C.v1_3_bigIntSetUnsignedBytes)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSetSignedBytes", v1_3_bigIntSetSignedBytes, C.v1_3_bigIntSetSignedBytes)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntIsInt64", v1_3_bigIntIsInt64, C.v1_3_bigIntIsInt64)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetInt64", v1_3_bigIntGetInt64, C.v1_3_bigIntGetInt64)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSetInt64", v1_3_bigIntSetInt64, C.v1_3_bigIntSetInt64)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntAdd", v1_3_bigIntAdd, C.v1_3_bigIntAdd)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSub", v1_3_bigIntSub, C.v1_3_bigIntSub)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntMul", v1_3_bigIntMul, C.v1_3_bigIntMul)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntTDiv", v1_3_bigIntTDiv, C.v1_3_bigIntTDiv)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntTMod", v1_3_bigIntTMod, C.v1_3_bigIntTMod)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntEDiv", v1_3_bigIntEDiv, C.v1_3_bigIntEDiv)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntEMod", v1_3_bigIntEMod, C.v1_3_bigIntEMod)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntAbs", v1_3_bigIntAbs, C.v1_3_bigIntAbs)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntNeg", v1_3_bigIntNeg, C.v1_3_bigIntNeg)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntSign", v1_3_bigIntSign, C.v1_3_bigIntSign)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntCmp", v1_3_bigIntCmp, C.v1_3_bigIntCmp)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntNot", v1_3_bigIntNot, C.v1_3_bigIntNot)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntAnd", v1_3_bigIntAnd, C.v1_3_bigIntAnd)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntOr", v1_3_bigIntOr, C.v1_3_bigIntOr)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntXor", v1_3_bigIntXor, C.v1_3_bigIntXor)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntShr", v1_3_bigIntShr, C.v1_3_bigIntShr)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntShl", v1_3_bigIntShl, C.v1_3_bigIntShl)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntFinishUnsigned", v1_3_bigIntFinishUnsigned, C.v1_3_bigIntFinishUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntFinishSigned", v1_3_bigIntFinishSigned, C.v1_3_bigIntFinishSigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntStorageStoreUnsigned", v1_3_bigIntStorageStoreUnsigned, C.v1_3_bigIntStorageStoreUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntStorageLoadUnsigned", v1_3_bigIntStorageLoadUnsigned, C.v1_3_bigIntStorageLoadUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetUnsignedArgument", v1_3_bigIntGetUnsignedArgument, C.v1_3_bigIntGetUnsignedArgument)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetSignedArgument", v1_3_bigIntGetSignedArgument, C.v1_3_bigIntGetSignedArgument)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetCallValue", v1_3_bigIntGetCallValue, C.v1_3_bigIntGetCallValue)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetDCDTCallValue", v1_3_bigIntGetDCDTCallValue, C.v1_3_bigIntGetDCDTCallValue)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetDCDTExternalBalance", v1_3_bigIntGetDCDTExternalBalance, C.v1_3_bigIntGetDCDTExternalBalance)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("bigIntGetExternalBalance", v1_3_bigIntGetExternalBalance, C.v1_3_bigIntGetExternalBalance)
	if err != nil {
		return nil, err
	}

	return imports, nil
}

const maxBigIntByteLenForNormalCost = 32

func useExtraGasForOperations(metering vmhost.MeteringContext, values []*big.Int) {
	for _, val := range values {
		byteLen := val.BitLen() / 8
		if byteLen > maxBigIntByteLenForNormalCost {
			metering.UseGas(math.MulUint64(uint64(byteLen), metering.GasSchedule().BaseOperationCost.DataCopyPerByte))
		}
	}
}

//export v1_3_bigIntGetUnsignedArgument
func v1_3_bigIntGetUnsignedArgument(context unsafe.Pointer, id int32, destination int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetUnsignedArgument
	metering.UseGas(gasToUse)

	args := runtime.Arguments()
	if int32(len(args)) <= id {
		return
	}

	value := bigInt.GetOne(destination)

	value.SetBytes(args[id])
}

//export v1_3_bigIntGetSignedArgument
func v1_3_bigIntGetSignedArgument(context unsafe.Pointer, id int32, destination int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetSignedArgument
	metering.UseGas(gasToUse)

	args := runtime.Arguments()
	if int32(len(args)) <= id {
		return
	}

	value := bigInt.GetOne(destination)

	twos.SetBytes(value, args[id])
}

//export v1_3_bigIntStorageStoreUnsigned
func v1_3_bigIntStorageStoreUnsigned(context unsafe.Pointer, keyOffset int32, keyLength int32, source int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntStorageStoreUnsigned
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return 0
	}

	value := bigInt.GetOne(source)
	bytes := value.Bytes()

	storageStatus, err := storage.SetStorage(key, bytes)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(storageStatus)
}

//export v1_3_bigIntStorageLoadUnsigned
func v1_3_bigIntStorageLoadUnsigned(context unsafe.Pointer, keyOffset int32, keyLength int32, destination int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntStorageLoadUnsigned
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return 0
	}

	bytes := storage.GetStorage(key)

	value := bigInt.GetOne(destination)
	value.SetBytes(bytes)

	return int32(len(bytes))
}

//export v1_3_bigIntGetCallValue
func v1_3_bigIntGetCallValue(context unsafe.Pointer, destination int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetCallValue
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(destination)
	value.Set(runtime.GetVMInput().CallValue)
}

//export v1_3_bigIntGetDCDTCallValue
func v1_3_bigIntGetDCDTCallValue(context unsafe.Pointer, destination int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetCallValue
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(destination)
	dcdtTransfer := getFirstDCDTTransferIfExist(runtime.GetVMInput())
	value.Set(dcdtTransfer.DCDTValue)
}

//export v1_3_bigIntGetExternalBalance
func v1_3_bigIntGetExternalBalance(context unsafe.Pointer, addressOffset int32, result int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	blockchain := vmhost.GetBlockchainContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetExternalBalance
	metering.UseGas(gasToUse)

	address, err := runtime.MemLoad(addressOffset, vmhost.AddressLen)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return
	}

	balance := blockchain.GetBalance(address)
	value := bigInt.GetOne(result)

	value.SetBytes(balance)
}

//export v1_3_bigIntGetDCDTExternalBalance
func v1_3_bigIntGetDCDTExternalBalance(context unsafe.Pointer, addressOffset int32, tokenIDOffset int32, tokenIDLen int32, nonce int64, result int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetExternalBalance
	metering.UseGas(gasToUse)

	dcdtData, err := getDCDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return
	}
	if dcdtData == nil {
		return
	}

	value := bigInt.GetOne(result)
	value.Set(dcdtData.Value)
}

//export v1_3_bigIntNew
func v1_3_bigIntNew(context unsafe.Pointer, smallValue int64) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntNew
	metering.UseGas(gasToUse)

	return bigInt.Put(smallValue)
}

//export v1_3_bigIntUnsignedByteLength
func v1_3_bigIntUnsignedByteLength(context unsafe.Pointer, reference int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntUnsignedByteLength
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)

	bytes := value.Bytes()
	return int32(len(bytes))
}

//export v1_3_bigIntSignedByteLength
func v1_3_bigIntSignedByteLength(context unsafe.Pointer, reference int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSignedByteLength
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)

	bytes := twos.ToBytes(value)
	return int32(len(bytes))
}

//export v1_3_bigIntGetUnsignedBytes
func v1_3_bigIntGetUnsignedBytes(context unsafe.Pointer, reference int32, byteOffset int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetUnsignedBytes
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)
	bytes := value.Bytes()

	err := runtime.MemStore(byteOffset, bytes)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return 0
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(len(bytes)))
	metering.UseGas(gasToUse)

	return int32(len(bytes))
}

//export v1_3_bigIntGetSignedBytes
func v1_3_bigIntGetSignedBytes(context unsafe.Pointer, reference int32, byteOffset int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetSignedBytes
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)
	bytes := twos.ToBytes(value)

	err := runtime.MemStore(byteOffset, bytes)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return 0
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(len(bytes)))
	metering.UseGas(gasToUse)

	return int32(len(bytes))
}

//export v1_3_bigIntSetUnsignedBytes
func v1_3_bigIntSetUnsignedBytes(context unsafe.Pointer, destination int32, byteOffset int32, byteLength int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSetUnsignedBytes
	metering.UseGas(gasToUse)

	bytes, err := runtime.MemLoad(byteOffset, byteLength)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return
	}

	value := bigInt.GetOne(destination)
	value.SetBytes(bytes)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(len(bytes)))
	metering.UseGas(gasToUse)
}

//export v1_3_bigIntSetSignedBytes
func v1_3_bigIntSetSignedBytes(context unsafe.Pointer, destination int32, byteOffset int32, byteLength int32) {
	bigInt := vmhost.GetBigIntContext(context)
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSetSignedBytes
	metering.UseGas(gasToUse)

	bytes, err := runtime.MemLoad(byteOffset, byteLength)
	if vmhost.WithFault(err, context, runtime.BigIntAPIErrorShouldFailExecution()) {
		return
	}

	value := bigInt.GetOne(destination)
	twos.SetBytes(value, bytes)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(len(bytes)))
	metering.UseGas(gasToUse)
}

//export v1_3_bigIntIsInt64
func v1_3_bigIntIsInt64(context unsafe.Pointer, handle int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntIsInt64
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(handle)
	if value.IsInt64() {
		return 1
	}
	return 0
}

//export v1_3_bigIntGetInt64
func v1_3_bigIntGetInt64(context unsafe.Pointer, handle int32) int64 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntGetInt64
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(handle)
	return value.Int64()
}

//export v1_3_bigIntSetInt64
func v1_3_bigIntSetInt64(context unsafe.Pointer, destination int32, value int64) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest := bigInt.GetOne(destination)
	dest.SetInt64(value)
}

//export v1_3_bigIntAdd
func v1_3_bigIntAdd(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	dest.Add(a, b)
}

//export v1_3_bigIntSub
func v1_3_bigIntSub(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	dest.Sub(a, b)
}

//export v1_3_bigIntMul
func v1_3_bigIntMul(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntMul
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	dest.Mul(a, b)
}

//export v1_3_bigIntTDiv
func v1_3_bigIntTDiv(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntTDiv
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	if b.Sign() == 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrDivZero, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Quo(a, b) // Quo implements truncated division (like Go)
}

//export v1_3_bigIntTMod
func v1_3_bigIntTMod(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	if b.Sign() == 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrDivZero, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Rem(a, b) // Rem implements truncated modulus (like Go)
}

//export v1_3_bigIntEDiv
func v1_3_bigIntEDiv(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	if b.Sign() == 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrDivZero, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Div(a, b) // Div implements Euclidean division (unlike Go)
}

//export v1_3_bigIntEMod
func v1_3_bigIntEMod(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{dest, a, b})
	if b.Sign() == 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrDivZero, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Mod(a, b) // Mod implements Euclidean division (unlike Go)
}

//export v1_3_bigIntAbs
func v1_3_bigIntAbs(context unsafe.Pointer, destination, op int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a := bigInt.GetTwo(destination, op)
	useExtraGasForOperations(metering, []*big.Int{dest, a})
	dest.Abs(a)
}

//export v1_3_bigIntNeg
func v1_3_bigIntNeg(context unsafe.Pointer, destination, op int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a := bigInt.GetTwo(destination, op)
	useExtraGasForOperations(metering, []*big.Int{dest, a})
	dest.Neg(a)
}

//export v1_3_bigIntSign
func v1_3_bigIntSign(context unsafe.Pointer, op int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSign
	metering.UseGas(gasToUse)

	a := bigInt.GetOne(op)
	useExtraGasForOperations(metering, []*big.Int{a})
	return int32(a.Sign())
}

//export v1_3_bigIntCmp
func v1_3_bigIntCmp(context unsafe.Pointer, op1, op2 int32) int32 {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntCmp
	metering.UseGas(gasToUse)

	a, b := bigInt.GetTwo(op1, op2)
	useExtraGasForOperations(metering, []*big.Int{a, b})
	return int32(a.Cmp(b))
}

//export v1_3_bigIntNot
func v1_3_bigIntNot(context unsafe.Pointer, destination, op int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a := bigInt.GetTwo(destination, op)
	useExtraGasForOperations(metering, []*big.Int{dest, a})
	if a.Sign() < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrBitwiseNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Not(a)
}

//export v1_3_bigIntAnd
func v1_3_bigIntAnd(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{a, b})
	if a.Sign() < 0 || b.Sign() < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrBitwiseNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.And(a, b)
}

//export v1_3_bigIntOr
func v1_3_bigIntOr(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{a, b})
	if a.Sign() < 0 || b.Sign() < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrBitwiseNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Or(a, b)
}

//export v1_3_bigIntXor
func v1_3_bigIntXor(context unsafe.Pointer, destination, op1, op2 int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a, b := bigInt.GetThree(destination, op1, op2)
	useExtraGasForOperations(metering, []*big.Int{a, b})
	if a.Sign() < 0 || b.Sign() < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrBitwiseNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Xor(a, b)
}

//export v1_3_bigIntShr
func v1_3_bigIntShr(context unsafe.Pointer, destination, op, bits int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a := bigInt.GetTwo(destination, op)
	useExtraGasForOperations(metering, []*big.Int{a})
	if a.Sign() < 0 || bits < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrShiftNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Rsh(a, uint(bits))
	useExtraGasForOperations(metering, []*big.Int{dest})
}

//export v1_3_bigIntShl
func v1_3_bigIntShl(context unsafe.Pointer, destination, op, bits int32) {
	bigInt := vmhost.GetBigIntContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntSub
	metering.UseGas(gasToUse)

	dest, a := bigInt.GetTwo(destination, op)
	useExtraGasForOperations(metering, []*big.Int{a})
	if a.Sign() < 0 || bits < 0 {
		runtime := vmhost.GetRuntimeContext(context)
		vmhost.WithFault(vmhost.ErrShiftNegative, context, runtime.BigIntAPIErrorShouldFailExecution())
		return
	}
	dest.Lsh(a, uint(bits))
	useExtraGasForOperations(metering, []*big.Int{dest})
}

//export v1_3_bigIntFinishUnsigned
func v1_3_bigIntFinishUnsigned(context unsafe.Pointer, reference int32) {
	bigInt := vmhost.GetBigIntContext(context)
	output := vmhost.GetOutputContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntFinishUnsigned
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)
	bigIntBytes := value.Bytes()
	output.Finish(bigIntBytes)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(len(value.Bytes())))
	metering.UseGas(gasToUse)
}

//export v1_3_bigIntFinishSigned
func v1_3_bigIntFinishSigned(context unsafe.Pointer, reference int32) {
	bigInt := vmhost.GetBigIntContext(context)
	output := vmhost.GetOutputContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BigIntAPICost.BigIntFinishSigned
	metering.UseGas(gasToUse)

	value := bigInt.GetOne(reference)
	bigInt2cBytes := twos.ToBytes(value)
	output.Finish(bigInt2cBytes)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(len(bigInt2cBytes)))
	metering.UseGas(gasToUse)
}
