package svm

import "unsafe"

const (
	AddressLength int = 20
	TxIdLength    int = 32
	StateLength   int = 32
	AmountLength  int = 8
	TxNonceLength int = 16
	GasLength     int = 8
	GasFeeLength  int = 8
	LayerLength   int = 8
)

// Declaring types aliases used throughout the project.
type TxType uint8
type Amount uint64
type Address [AddressLength]byte
type TemplateAddr [AddressLength]byte
type TxId [TxIdLength]byte
type State [StateLength]byte
type Gas uint64
type GasFee uint64
type Layer uint64
type Log []byte

// `Runtime` wraps the raw-Runtime returned by SVM C-API
type Runtime struct {
	raw unsafe.Pointer
}

// Holds the currently executed `Node Context`.
// Addionally, contains data implied/computed from the `input` transaction.
type Context struct {
	Layer Layer
	TxId  TxId
}

// Encapsulates a `Transaction Nonce`. (Since `Golang` has no `unit128` primitive out-of-the-box).
//
// Used for implementing the `Nonce Scheme` implemented within the `Template` associated with the `Account`.
type TxNonce struct {
	Upper uint64
	Lower uint64
}

// A `Transaction Type` enum
const (
	DeployType TxType = 0
	SpawnType  TxType = 1
	CallType   TxType = 2
)

// Holds the `Envelope` of a transaction.
type Envelope struct {
	Type      TxType
	Principal Address
	Amount    Amount
	TxNonce   TxNonce
	GasLimit  Gas
	GasFee    GasFee
}

// Holds an `Account` basic information.
type Account struct {
	Addr    Address
	Balance Amount
	Counter TxNonce
}

// Holds the data returned after executing a `Deploy` transaction.
type DeployReceipt struct {
	Success      bool
	Error        RuntimeError
	TemplateAddr TemplateAddr
	GasUsed      Gas
	Logs         []Log
}

// Holds the data returned after executing a `Spawn` transaction.
type SpawnReceipt struct {
	Success         bool
	Error           RuntimeError
	AccountAddr     Address
	InitState       State
	GasUsed         Gas
	Logs            []Log
	TouchedAccounts []Address
}

// Holds the data returned after executing a `Call` transaction.
type CallReceipt struct {
	Success         bool
	Error           RuntimeError
	NewState        State
	ReturnData      []byte
	GasUsed         Gas
	Logs            []Log
	TouchedAccounts []Address
}
