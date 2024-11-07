package vmhooks

// // Declare the function signatures (see [cgo](https://golang.org/cmd/cgo/)).
//
// #include <stdlib.h>
// typedef unsigned char uint8_t;
// typedef int int32_t;
//
// extern long long v1_3_smallIntGetUnsignedArgument(void *context, int32_t id);
// extern long long v1_3_smallIntGetSignedArgument(void *context, int32_t id);
//
// extern void			v1_3_smallIntFinishUnsigned(void* context, long long value);
// extern void			v1_3_smallIntFinishSigned(void* context, long long value);
//
// extern int32_t		v1_3_smallIntStorageStoreUnsigned(void *context, int32_t keyOffset, int32_t keyLength, long long value);
// extern int32_t		v1_3_smallIntStorageStoreSigned(void *context, int32_t keyOffset, int32_t keyLength, long long value);
// extern long long v1_3_smallIntStorageLoadUnsigned(void *context, int32_t keyOffset, int32_t keyLength);
// extern long long v1_3_smallIntStorageLoadSigned(void *context, int32_t keyOffset, int32_t keyLength);
//
// extern long long v1_3_int64getArgument(void *context, int32_t id);
// extern int32_t		v1_3_int64storageStore(void *context, int32_t keyOffset, int32_t keyLength , long long value);
// extern long long v1_3_int64storageLoad(void *context, int32_t keyOffset, int32_t keyLength );
// extern void			v1_3_int64finish(void* context, long long value);
//
import "C"

import (
	"math/big"
	"unsafe"

	"github.com/kalyan3104/k-chain-vm-v1_3-go/vmhost"
	"github.com/kalyan3104/k-chain-vm-v1_3-go/wasmer"
	twos "github.com/kalyan3104/k-components-big-int/twos-complement"
)

// SmallIntImports creates a new wasmer.Imports populated with the small int (int64/uint64) API methods
func SmallIntImports(imports *wasmer.Imports) (*wasmer.Imports, error) {
	imports = imports.Namespace("env")

	imports, err := imports.Append("smallIntGetUnsignedArgument", v1_3_smallIntGetUnsignedArgument, C.v1_3_smallIntGetUnsignedArgument)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntGetSignedArgument", v1_3_smallIntGetSignedArgument, C.v1_3_smallIntGetSignedArgument)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntFinishUnsigned", v1_3_smallIntFinishUnsigned, C.v1_3_smallIntFinishUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntFinishSigned", v1_3_smallIntFinishSigned, C.v1_3_smallIntFinishSigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntStorageStoreUnsigned", v1_3_smallIntStorageStoreUnsigned, C.v1_3_smallIntStorageStoreUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntStorageStoreSigned", v1_3_smallIntStorageStoreSigned, C.v1_3_smallIntStorageStoreSigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntStorageLoadUnsigned", v1_3_smallIntStorageLoadUnsigned, C.v1_3_smallIntStorageLoadUnsigned)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("smallIntStorageLoadSigned", v1_3_smallIntStorageLoadSigned, C.v1_3_smallIntStorageLoadSigned)
	if err != nil {
		return nil, err
	}

	// the last are just for backwards compatibility:

	imports, err = imports.Append("int64getArgument", v1_3_int64getArgument, C.v1_3_int64getArgument)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("int64storageStore", v1_3_int64storageStore, C.v1_3_int64storageStore)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("int64storageLoad", v1_3_int64storageLoad, C.v1_3_int64storageLoad)
	if err != nil {
		return nil, err
	}

	imports, err = imports.Append("int64finish", v1_3_int64finish, C.v1_3_int64finish)
	if err != nil {
		return nil, err
	}

	return imports, nil
}

//export v1_3_smallIntGetUnsignedArgument
func v1_3_smallIntGetUnsignedArgument(context unsafe.Pointer, id int32) int64 {
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64GetArgument
	metering.UseGas(gasToUse)

	args := runtime.Arguments()
	if id < 0 || id >= int32(len(args)) {
		vmhost.WithFault(vmhost.ErrArgIndexOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}

	arg := args[id]
	argBigInt := big.NewInt(0).SetBytes(arg)
	if !argBigInt.IsUint64() {
		vmhost.WithFault(vmhost.ErrArgOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}
	return int64(argBigInt.Uint64())
}

//export v1_3_smallIntGetSignedArgument
func v1_3_smallIntGetSignedArgument(context unsafe.Pointer, id int32) int64 {
	runtime := vmhost.GetRuntimeContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64GetArgument
	metering.UseGas(gasToUse)

	args := runtime.Arguments()
	if id < 0 || id >= int32(len(args)) {
		vmhost.WithFault(vmhost.ErrArgIndexOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}

	arg := args[id]
	argBigInt := twos.SetBytes(big.NewInt(0), arg)
	if !argBigInt.IsInt64() {
		vmhost.WithFault(vmhost.ErrArgOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}
	return argBigInt.Int64()
}

//export v1_3_smallIntFinishUnsigned
func v1_3_smallIntFinishUnsigned(context unsafe.Pointer, value int64) {
	output := vmhost.GetOutputContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64Finish
	metering.UseGas(gasToUse)

	valueBytes := big.NewInt(0).SetUint64(uint64(value)).Bytes()
	output.Finish(valueBytes)
}

//export v1_3_smallIntFinishSigned
func v1_3_smallIntFinishSigned(context unsafe.Pointer, value int64) {
	output := vmhost.GetOutputContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64Finish
	metering.UseGas(gasToUse)

	valueBytes := twos.ToBytes(big.NewInt(value))
	output.Finish(valueBytes)
}

//export v1_3_smallIntStorageStoreUnsigned
func v1_3_smallIntStorageStoreUnsigned(context unsafe.Pointer, keyOffset int32, keyLength int32, value int64) int32 {
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64StorageStore
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return -1
	}

	valueBytes := big.NewInt(0).SetUint64(uint64(value)).Bytes()
	storageStatus, err := storage.SetStorage(key, valueBytes)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return -1
	}

	return int32(storageStatus)
}

//export v1_3_smallIntStorageStoreSigned
func v1_3_smallIntStorageStoreSigned(context unsafe.Pointer, keyOffset int32, keyLength int32, value int64) int32 {
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64StorageStore
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return -1
	}

	valueBytes := twos.ToBytes(big.NewInt(value))
	storageStatus, err := storage.SetStorage(key, valueBytes)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return -1
	}

	return int32(storageStatus)
}

//export v1_3_smallIntStorageLoadUnsigned
func v1_3_smallIntStorageLoadUnsigned(context unsafe.Pointer, keyOffset int32, keyLength int32) int64 {
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64StorageLoad
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return 0
	}

	data := storage.GetStorage(key)
	valueBigInt := big.NewInt(0).SetBytes(data)
	if !valueBigInt.IsUint64() {
		vmhost.WithFault(vmhost.ErrStorageValueOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}

	return int64(valueBigInt.Uint64())
}

//export v1_3_smallIntStorageLoadSigned
func v1_3_smallIntStorageLoadSigned(context unsafe.Pointer, keyOffset int32, keyLength int32) int64 {
	runtime := vmhost.GetRuntimeContext(context)
	storage := vmhost.GetStorageContext(context)
	metering := vmhost.GetMeteringContext(context)

	gasToUse := metering.GasSchedule().BaseOpsAPICost.Int64StorageLoad
	metering.UseGas(gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if vmhost.WithFault(err, context, runtime.BaseOpsErrorShouldFailExecution()) {
		return 0
	}

	data := storage.GetStorage(key)
	valueBigInt := twos.SetBytes(big.NewInt(0), data)
	if !valueBigInt.IsInt64() {
		vmhost.WithFault(vmhost.ErrStorageValueOutOfRange, context, runtime.BaseOpsErrorShouldFailExecution())
		return 0
	}

	return valueBigInt.Int64()
}

//export v1_3_int64getArgument
func v1_3_int64getArgument(context unsafe.Pointer, id int32) int64 {
	// backwards compatibility
	return v1_3_smallIntGetSignedArgument(context, id)
}

//export v1_3_int64finish
func v1_3_int64finish(context unsafe.Pointer, value int64) {
	// backwards compatibility
	v1_3_smallIntFinishSigned(context, value)
}

//export v1_3_int64storageStore
func v1_3_int64storageStore(context unsafe.Pointer, keyOffset int32, keyLength int32, value int64) int32 {
	// backwards compatibility
	return v1_3_smallIntStorageStoreUnsigned(context, keyOffset, keyLength, value)
}

//export v1_3_int64storageLoad
func v1_3_int64storageLoad(context unsafe.Pointer, keyOffset int32, keyLength int32) int64 {
	// backwards compatibility
	return v1_3_smallIntStorageLoadUnsigned(context, keyOffset, keyLength)
}
