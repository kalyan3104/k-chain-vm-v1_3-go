#include "../../drtvm/context.h"
#include "../../drtvm/test_utils.h"
#include "../../drtvm/args.h"

byte executeValue[] = {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0};
byte self[32] = "\0\0\0\0\0\0\0\0\x0f\x0f" "parentSC..............";
byte vaultSC[] = "\0\0\0\0\0\0\0\0\x0F\x0F" "vaultSC...............";
byte DCDTTransfer[] = "DCDTTransfer";

void basic_transfer() {
	byte tokenName[265] = {0};
	int tokenNameLen = getDCDTTokenName(tokenName);

	byte dcdtValue[32] = {0};
	int dcdtValueLen = getDCDTValue(dcdtValue);

	dcdtValue[31] -= 1;

	BinaryArgs args = NewBinaryArgs();

	int lastArg = 0;
	lastArg = AddBinaryArg(&args, tokenName, tokenNameLen);
	lastArg = AddBinaryArg(&args, dcdtValue, dcdtValueLen);
	TrimLeftZeros(&args, lastArg);

	byte arguments[100];
	int argsLen = SerializeBinaryArgs(&args, arguments);

	int result = executeOnDestContext(
			1000000,
			self,
			executeValue,
			DCDTTransfer,
			sizeof DCDTTransfer - 1,
			args.numArgs,
		  (byte*)args.lengthsAsI32,
			args.serialized
	);
}
