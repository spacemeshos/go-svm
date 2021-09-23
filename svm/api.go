package svm

/*
#cgo LDFLAGS: -L ${SRCDIR}/../artifacts -lsvm
#include "svm.h"
#include "memory.h"
*/
import "C"
import (
	"unsafe"
)

// TODO: we might want to guard calling `Init` with a Mutex.
var initialized = false

type StateDigest = [32]byte

// `Init` should be called exactly once before interacting with any other API of SVM.
// Each future call to `NewRuntime` (see later) assumes the settings given the `Init` call.
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
//
//
// # Panics
//
// Panics when `Init` has already been called.
func Init(inMemory bool, path string) error {
	if initialized {
		panic("`Init` can be called only once")
	}

	bytes := ([]byte)(path)
	rawPath := (*C.uchar)(unsafe.Pointer(&bytes[0]))
	pathLen := (C.uint32_t)(uint32(len(path)))
	C.svm_init((C.bool)(inMemory), rawPath, pathLen)

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
	C.svm_runtime_create(&rt.raw)

	// TODO: add common functionality `svm_result_t`
	return rt, nil
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
	C.svm_runtime_destroy(rt.raw)
}

// Validates the `Deploy Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateDeploy(msg []byte) (bool, error) {
	rawMsg := (*C.uchar)(unsafe.Pointer(&msg[0]))
	msgLen := (C.uint32_t)(uint32(len(msg)))
	C.svm_validate_deploy(rt.raw, rawMsg, msgLen)

	return false, nil
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
func (rt *Runtime) Deploy(env Envelope, msg []byte, ctx Context) DeployReceipt {
	panic("TODO")
}

// Validates the `Spawn Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateSpawn(msg []byte) (bool, ValidateError) {
	panic("TODO")
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
func (rt *Runtime) ValidateCall(msg []byte) (bool, ValidateError) {
	panic("TODO")
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
func (rt *Runtime) Rewind(layer Layer) (StateDigest, error) {
	panic("TODO")
}

// Commits the dirty changes of `SVM` and signals the termination of the current layer.
// On success returns `(layer, nil)` when `layer` is the value of the previous `current layer`.
// In other words, returns the `layer` associated with the just-committed changes.
//
// In case commits fails (for example, persisting to disk failure) - returns `(0, error)`
func (rt *Runtime) Commit() (Layer, StateDigest, error) {
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
func (rt *Runtime) IncreaseBalance(addr Account, amount Amount) {
	panic("TODO")
}
