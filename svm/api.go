package svm

/*
#cgo CFLAGS: -I.
#cgo linux LDFLAGS: ${SRCDIR}/artifacts/bins-Linux-release/libsvm.a -lm -ldl
#cgo darwin LDFLAGS: ${SRCDIR}/artifacts/bins-macOS-release/libsvm.a -lm -ldl -framework Security -framework Foundation
#cgo windows LDFLAGS: -L ${SRCDIR}/artifacts/bins-Windows-release/ -lsvm -lm -ldl
#include "svm.h"
#include "memory.h"
*/
import "C"
import (
	"encoding/binary"
	"errors"
	"unsafe"
)

func copySvmResult(res C.struct_svm_result_t) ([]byte, error) {
	size := C.int(res.buf_size)

	receipt := ([]byte)(nil)
	err := (error)(nil)

	if res.receipt != nil {
		ptr := unsafe.Pointer(res.receipt)
		receipt = C.GoBytes(ptr, size)
	} else if res.error != nil {
		ptr := unsafe.Pointer(res.error)
		err = errors.New(string(C.GoBytes(ptr, size)))
	}

	C.free(unsafe.Pointer(res.receipt))
	C.free(unsafe.Pointer(res.error))

	return receipt, err
}

// TODO: we might want to guard calling `Init` with a Mutex.
var initialized = false

// `Init` should be called at least once before interacting with any other API of SVM.
// Each future call to `NewRuntime` (see later) assumes the settings given the `Init` call.
//
// Please note that this function is idempotent and won't do anything after the
// first call.
//
// # Params
//
// * `inMemory` - whether the data of the `SVM Global State` will be in-memory or persisted.
// * `path` 	- the path under which `SVM Global State` will store its content.
//   This is relevant only when `isMemory=false` (otherwise the `path` value will be ignored).
//
// # Returns
//
// Returns an error in case the initialization has failed.
func Init(inMemory bool, path string) error {
	if initialized {
		return nil
	}

	bytes := ([]byte)(path)
	rawPath := (*C.uchar)(unsafe.Pointer(&bytes))
	pathLen := (C.uint32_t)(uint32(len(path)))
	var res C.struct_svm_result_t = C.svm_init((C.bool)(inMemory), rawPath, pathLen)
	copySvmResult(res)

	return nil
}

// Asserts that `Init` has already been called.
//
// # Panics
//
// Panics when `Init` has **NOT** been previously called.
func AssertInitialized() {
	if !initialized {
		panic("Forgot to call `Init`")
	}
}

// `NewRuntime` creates a new `Runtime`.
//
// On success returns it and the `error` is set to `nil`.
// On failure returns `(nil, error).
func NewRuntime() (*Runtime, error) {
	rt := &Runtime{}
	res := C.svm_runtime_create(&rt.raw)
	_, err := copySvmResult(res)

	return rt, err
}

func RuntimesCount() int {
	count := C.uint64_t(0)
	C.svm_runtimes_count(&count)

	return int(count)
}

func ReceiptsCount() int {
	return 0
	// count := C.uint32(0)
	// result := C.svm_receipts_count(&count)
	// return int(count)
}

// Releases the SVM Runtime
func (rt *Runtime) Destroy() {
	if rt.raw != nil {
		C.svm_runtime_destroy(rt.raw)
	}
}

// Validates the `Deploy Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateDeploy(msg []byte) (bool, error) {
	rawMsg := (*C.uchar)(unsafe.Pointer(&msg))
	msgLen := (C.uint32_t)(len(msg))
	res := C.svm_validate_deploy(rt.raw, rawMsg, msgLen)
	_, err := copySvmResult(res)

	return err == nil, err
}

// Executes a `Deploy` transaction and returns back a receipt.
//
// # Params
//
// * `env` - the transaction `Envelope`
// * `msg` - the transaction `Message`
// * `ctx` - the executed `Context` (the `current layer` etc).
//
// # Notes
//
// A Receipt is always being returned, even if there was an internal error inside SVM.
func (rt *Runtime) Deploy(env *Envelope, msg []byte, ctx *Context) DeployReceipt {
	// `Envelope`
	envBytes := EncodeEnvelope(env)
	envPtr := (*C.uchar)(unsafe.Pointer(&envBytes))

	// `Message`
	msgPtr := (*C.uchar)(unsafe.Pointer(&msg))
	msgLen := (C.uint32_t)(uint32(len(msg)))

	// `Context`
	ctxBytes := EncodeContext(ctx)
	ctxPtr := (*C.uchar)(unsafe.Pointer(&ctxBytes))

	res := C.svm_deploy(rt.raw, envPtr, msgPtr, msgLen, ctxPtr)
	_, err := copySvmResult(res)
	if err != nil {
		panic(err)
	}

	// TODO
	return DeployReceipt{
		Success:      true,
		Error:        RuntimeError{},
		TemplateAddr: TemplateAddr{0},
		GasUsed:      Gas(0),
		Logs:         make([]Log, 0),
	}
}

// Validates the `Spawn Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateSpawn(msg []byte) (bool, *ValidateError) {
	rawMsg := (*C.uchar)(unsafe.Pointer(&msg))
	msgLen := (C.uint32_t)(len(msg))
	res := C.svm_validate_spawn(rt.raw, rawMsg, msgLen)
	_, err := copySvmResult(res)

	if err == nil {
		return true, nil
	} else {
		return false, &ValidateError{
			Kind:    ParseError,
			Message: err.Error(),
		}
	}
}

// Executes a `Spawn` transaction and returns back a receipt.
//
// # Params
//
// * `env` - the transaction `Envelope`
// * `msg` - the transaction `Message`
// * `ctx` - the executed `Context` (the `current layer` etc).
//
//
// # Notes
//
// A Receipt is always being returned, even if there was an internal error inside SVM.
func (rt *Runtime) Spawn(env Envelope, msg []byte, ctx Context) SpawnReceipt {
	panic("TODO")
}

// Validates the `Call Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateCall(msg []byte) (bool, *ValidateError) {
	rawMsg := (*C.uchar)(unsafe.Pointer(&msg))
	msgLen := (C.uint32_t)(len(msg))
	res := C.svm_validate_call(rt.raw, rawMsg, msgLen)
	_, err := copySvmResult(res)

	if err == nil {
		return true, nil
	} else {
		return false, &ValidateError{
			Kind:    ParseError,
			Message: err.Error(),
		}
	}
}

// Executes a `Call` transaction and returns back a receipt.
//
// # Params
//
// * `env` - the transaction `Envelope`
// * `msg` - the transaction `Message`
// * `ctx` - the executed `Context` (the `current layer` etc).
//
// # Notes
//
// A Receipt is always being returned, even if there was an internal error inside SVM.
func (rt *Runtime) Call(env Envelope, msg []byte, ctx Context) CallReceipt {
	panic("TODO")
}

// Executes the `Verify` stage and returns back a receipt.
//
// Calling `Verify` is in many ways very similar to the latter `Call` execution.
// For that reason, the `Verify` returns a `CallReceipt` as well.
//
// # Params
//
// * `env` - the transaction `Envelope`
// * `msg` - the transaction `Message`
// * `ctx` - the executed `Context` (the `current layer` etc).
//
// # Notes
//
// A Receipt is always being returned, even if there was an internal error inside SVM.
func (rt *Runtime) Verify(env Envelope, msg []byte, ctx Context) CallReceipt {
	panic("TODO")
}

// Signaling `SVM` that we are about to start playing a list of transactions under the input `layer` Layer.
//
// # Notes
//
// * The value of `layer` is expected to equal the last known `committed/rewinded` layer plus one
//   Any other `layer` given as input will result in an error returned.
//
// * Calling `Open` twice in a row will result in an `error` returned.
func (rt *Runtime) Open(layer Layer) error {
	panic("TODO")
}

// Rewinds the `SVM Global State` back to the input `layer`.
//
// In case there is no such layer to rewind to - returns an `error`.
func (rt *Runtime) Rewind(layer Layer) (State, error) {
	panic("TODO")
}

// Commits the dirty changes of `SVM` and signals the termination of the current layer.
// On success returns `(layer, nil)` when `layer` is the value of the previous `current layer`.
// In other words, returns the `layer` associated with the just-committed changes.
//
// In case commits fails (for example, persisting to disk failure) - returns `(0, error)`
func (rt *Runtime) Commit() (Layer, State, error) {
	panic("TODO")
}

// Given an `Account Address` - retrieves its most basic information encapuslated within an `Account` struct.
//
// Returns a `(nil, error)` in case the requested `Account` doesn't exist.
func (rt *Runtime) GetAccount(addr Address) (Account, error) {
	panic("TODO")
}

// Increases the balance of an Account (i.e printing coins)
//
// # Params
//
// * `addr`   - The `Account Address` we want to increase its balance.
// * `amount` - The `Amount` by which we are going to increase the account's balance.
func (rt *Runtime) IncreaseBalance(addr Address, amount Amount) {
	panic("TODO")
}

func NewEvelope(principal Address, amount Amount, txNonce TxNonce, gasLimit Gas, gasFee GasFee) Envelope {
	env := Envelope{}
	env.Principal = principal
	env.Amount = amount
	env.TxNonce = txNonce
	env.GasLimit = gasLimit
	env.GasFee = gasFee
	return env
}

//
// In other words, holds fields which are part of any transaction regardless of its type (i.e `Deploy/Spawn/Call`).
/// Encoding of a binary [`Envelope`].
///
/// ```text
///  +-------------+--------------+----------------+----------------+----------------+
///  |             |              |			       |                |				 |
///  |  Principal  |    Amount    |    Tx Nonce    |   Gas Limit    |    Gas Fee 	 |
///  |  (Address)  |    (u64)     |     (u128)     |     (u64)      |     (u64)	 	 |
///  |             |              |                |                |				 |
///  |  20 bytes   |   8 bytes    |    16 bytes    |    8 bytes     |    8 bytes     |
///  |             | (Big-Endian) |  (Big-Endian)  |  (Big-Endian)  |  (Big-Endian)  |
///  |             |              |                |                |			     |
///  +-------------+--------------+----------------+----------------+----------------+
/// ```
func EncodeEnvelope(env *Envelope) [EnvelopeLength]byte {
	bytes := [EnvelopeLength]byte{0}

	// `Principal`
	copy(bytes[:AddressLength], env.Principal[:])

	// `Amount`
	p := AddressLength
	binary.BigEndian.PutUint64(bytes[p:p+AmountLength], (uint64)(env.Amount))

	// `Tx Nonce`
	p += TxNonceLength
	binary.BigEndian.PutUint64(bytes[p:p+TxNonceLength/2], (uint64)(env.TxNonce.Upper))
	binary.BigEndian.PutUint64(bytes[p:p+TxNonceLength/2], (uint64)(env.TxNonce.Lower))

	// `Gas Limit`
	p += GasLength
	binary.BigEndian.PutUint64(bytes[p:p+GasLength], (uint64)(env.GasLimit))

	// `Gas Fee`
	p += GasFeeLength
	binary.BigEndian.PutUint64(bytes[p:p+GasFeeLength], (uint64)(env.GasFee))

	return bytes
}

/// ```text
///  +-------------+--------------+
///  |             |              |
///  |    Layer    |    Tx Id     |
///  |   (u64)     |    (Blob)    |
///  |             |              |
///  |  8 bytes    |   32 bytes   |
///  |             | 			  |
///  +-------------+--------------+
/// ```
func EncodeContext(ctx *Context) [ContextLength]byte {
	bytes := [ContextLength]byte{0}

	// `Layer`
	p := 0
	binary.BigEndian.PutUint64(bytes[:LayerLength], (uint64)(ctx.Layer))

	// `Tx Id`
	p += LayerLength
	copy(bytes[p:p+TxIdLength], ctx.TxId[:])

	return bytes
}
