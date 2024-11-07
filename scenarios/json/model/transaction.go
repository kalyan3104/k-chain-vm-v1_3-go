package scenjsonmodel

// TransactionType describes the type of simulate transaction
type TransactionType int

const (
	// ScDeploy describes a transaction that deploys a new contract
	ScDeploy TransactionType = iota

	// ScCall describes a regular smart contract call
	ScCall

	// ScQuery simulates an off-chain call.
	// It is like a SCCall, but without a sender and with infinite gas.
	ScQuery

	// Transfer is an MOA transfer transaction without calling a smart contract
	Transfer

	// ValidatorReward is when the protocol sends a validator reward to the target account.
	// It increases the balance, but also increments the reward value in storage.
	ValidatorReward
)

// HasSender is a helper function to indicate if transaction has `to` field.
func (tt TransactionType) HasSender() bool {
	return tt != ScQuery && tt != ValidatorReward
}

// HasReceiver is a helper function to indicate if transaction has receiver.
func (tt TransactionType) HasReceiver() bool {
	return tt != ScDeploy
}

// IsSmartContractTx indicates whether tx type allows an `expect` field.
func (tt TransactionType) IsSmartContractTx() bool {
	return tt == ScDeploy || tt == ScCall || tt == ScQuery
}

// HasValue indicates whether tx type allows a `value` field.
func (tt TransactionType) HasValue() bool {
	return tt != ScQuery
}

// HasDCDT is a helper function to indicate if transaction has `dcdtValue` or `dcdtToken` fields.
func (tt TransactionType) HasDCDT() bool {
	return tt != ScQuery && tt != ValidatorReward
}

// HasFunction indicates whether tx type allows a `function` field.
func (tt TransactionType) HasFunction() bool {
	return tt == ScCall || tt == ScQuery
}

// HasGas is a helper function to indicate if transaction has `dcdtValue` or `dcdtToken` fields.
func (tt TransactionType) HasGas() bool {
	return tt == ScDeploy || tt == ScCall || tt.HasDCDT()
}

// Transaction is a json object representing a transaction.
type Transaction struct {
	Type      TransactionType
	Nonce     JSONUint64
	Value     JSONBigInt
	DCDTValue *DCDTTxData
	From      JSONBytesFromString
	To        JSONBytesFromString
	Function  string
	Code      JSONBytesFromString
	Arguments []JSONBytesFromTree
	GasPrice  JSONUint64
	GasLimit  JSONUint64
}

// TransactionResult is a json object representing an expected transaction result.
type TransactionResult struct {
	Out             []JSONCheckBytes
	Status          JSONCheckBigInt
	Message         JSONCheckBytes
	Gas             JSONCheckUint64
	Refund          JSONCheckBigInt
	LogsStar        bool
	LogsUnspecified bool
	LogHash         string
	Logs            []*LogEntry
}

// LogEntry is a json object representing an expected transaction result log entry.
type LogEntry struct {
	Address    JSONCheckBytes
	Identifier JSONCheckBytes
	Topics     []JSONCheckBytes
	Data       JSONCheckBytes
}
