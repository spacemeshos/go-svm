package svm

// #include "svm.h"
// #include "memory.h"
import "C"
import (
	"errors"
	"log"
	"sync"
	"unsafe"
)

type svmParams struct {
	envPtr *C.uchar
	msgPtr *C.uchar
	msgLen C.uint32_t
	ctxPtr *C.uchar
}

var initialized = false
var initializedGuard = sync.Mutex{}

// Allows for creating new SVM runtime instances via NewRuntime.
type API struct{}

// Init is the entry point for interacting with SVM. It runs SVM initialization
// logic; it is fully thread-safe and idempotent.
func Init() (*API, error) {
	initializedGuard.Lock()
	defer initializedGuard.Unlock()

	initialized = true

	res := C.svm_init()
	if _, err := copySvmResult(res); err != nil {
		return nil, err
	}

	return &API{}, nil
}

// Asserts that `Init` has already been called.
//
// # Panics
//
// Panics when `Init` has **NOT** been previously called.
func AssertInitialized() {
	initializedGuard.Lock()
	defer initializedGuard.Unlock()

	if !initialized {
		panic("Forgot to call `Init`")
	}
}

// `NewRuntime` creates a new `Runtime`.
//
// # Params
//
// * `inMemory` - whether the data of the `SVM Global State` will be in-memory or persisted.
// * `path` 	- the path under which `SVM Global State` will store its content.
//   This is relevant only when `isMemory=false` (otherwise the `path` value will be ignored).
//
// On success returns it and the `error` is set to `nil`.
// On failure returns `(nil, error).
func (*API) NewRuntime(inMemory bool, path string) (*Runtime, error) {
	rt := &Runtime{}

	var res C.svm_result_t
	if inMemory {
		res = C.svm_runtime_create(&rt.raw, nil, 0)
	} else {
		bytes := ([]byte)(path)
		rawPath := (*C.uchar)(unsafe.Pointer(&bytes))
		pathLen := (C.uint32_t)(uint32(len(path)))
		res = C.svm_runtime_create(&rt.raw, rawPath, pathLen)
	}
	_, err := copySvmResult(res)

	return rt, err
}

func (*API) RuntimesCount() int {
	return int(C.svm_runtimes_count())
}

func (*API) ReceiptsCount() int {
	// TODO
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
	return runValidation(msg, func(rawMsg *C.uchar, msgLen C.uint32_t) C.svm_result_t {
		return C.svm_validate_deploy(rt.raw, rawMsg, msgLen)
	})
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
func (rt *Runtime) Deploy(env *Envelope, msg []byte, ctx *Context) (*DeployReceipt, error) {
	object, err := runAction(env, msg, ctx, func(params *svmParams) C.svm_result_t {
		return C.svm_deploy(rt.raw, params.envPtr, params.msgPtr, params.msgLen, params.ctxPtr)
	})

	if err != nil {
		return nil, err
	}

	return object.(*DeployReceipt), nil
}

// Validates the `Spawn Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateSpawn(msg []byte) (bool, error) {
	return runValidation(msg, func(rawMsg *C.uchar, msgLen C.uint32_t) C.svm_result_t {
		return C.svm_validate_spawn(rt.raw, rawMsg, msgLen)
	})
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
func (rt *Runtime) Spawn(env *Envelope, msg []byte, ctx *Context) (*SpawnReceipt, error) {
	object, err := runAction(env, msg, ctx, func(params *svmParams) C.svm_result_t {
		return C.svm_spawn(rt.raw, params.envPtr, params.msgPtr, params.msgLen, params.ctxPtr)
	})

	if err != nil {
		return nil, err
	}

	return object.(*SpawnReceipt), nil
}

// Validates the `Call Message` given in its binary form.
//
// Returns `(true, nil)` when the `msg` is syntactically valid,
// and `(false, error)` otherwise.  In that case `error` will have non-`nil` value.
func (rt *Runtime) ValidateCall(msg []byte) (bool, error) {
	return runValidation(msg, func(rawMsg *C.uchar, msgLen C.uint32_t) C.svm_result_t {
		return C.svm_validate_call(rt.raw, rawMsg, msgLen)
	})
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
func (rt *Runtime) Call(env *Envelope, msg []byte, ctx *Context) (*CallReceipt, error) {
	object, err := runAction(env, msg, ctx, func(params *svmParams) C.svm_result_t {
		return C.svm_call(rt.raw, params.envPtr, params.msgPtr, params.msgLen, params.ctxPtr)
	})

	if err != nil {
		return nil, err
	}

	return object.(*CallReceipt), nil
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
func (rt *Runtime) Verify(env *Envelope, msg []byte, ctx *Context) (*CallReceipt, error) {
	object, err := runAction(env, msg, ctx, func(params *svmParams) C.svm_result_t {
		return C.svm_verify(rt.raw, params.envPtr, params.msgPtr, params.msgLen, params.ctxPtr)
	})

	if err != nil {
		return nil, err
	}

	return object.(*CallReceipt), nil
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
	res := C.svm_uncommitted_changes(rt.raw)
	_, err := copySvmResult(res)
	if err != nil {
		return err
	}
	log.Print("Ready to play SVM transactions in a new layer.")
	return nil
}

// Rewinds the `SVM Global State` back to the input `layer`.
//
// In case there is no such layer to rewind to - returns an `error`.
func (rt *Runtime) Rewind(layer Layer) (State, error) {
	res := C.svm_rewind(rt.raw, C.uint64_t(layer))
	_, err := copySvmResult(res)
	if err != nil {
		return State{}, err
	}
	state, err := rt.StateHash()
	if err != nil {
		return State{}, err
	}
	return state, nil
}

func (rt *Runtime) layerInfo() (uint64, State, error) {
	state := State{}
	statePtr := (*C.uchar)(unsafe.Pointer(&state))
	layer := uint64(0)
	layerPtr := (*C.uint64_t)(unsafe.Pointer(&layer))
	res := C.svm_layer_info(rt.raw, statePtr, layerPtr)
	_, err := copySvmResult(res)
	if err != nil {
		return 0, State{}, err
	}
	return layer, state, nil
}

func (rt *Runtime) StateHash() (State, error) {
	_, state, err := rt.layerInfo()
	return state, err
}

// Commits the dirty changes of `SVM` and signals the termination of the current layer.
// On success returns `(layer, nil)` when `layer` is the value of the previous `current layer`.
// In other words, returns the `layer` associated with the just-committed changes.
//
// In case commits fails (for example, persisting to disk failure) - returns `(0, error)`
func (rt *Runtime) Commit() (Layer, State, error) {
	res := C.svm_commit(rt.raw)

	_, err := copySvmResult(res)
	if err != nil {
		return Layer(0), State{}, err
	}

	layer, hash, err := rt.layerInfo()
	return Layer(layer), hash, nil
}

// Given an `Account Address` - retrieves its most basic information encapuslated within an `Account` struct.
//
// Returns a `(nil, error)` in case the requested `Account` doesn't exist.
func (rt *Runtime) GetAccount(addr Address) (Account, error) {
	var account C.svm_account
	addrRaw := (*C.uchar)(unsafe.Pointer(&addr[0]))

	res := C.svm_get_account(rt.raw, addrRaw, &account)

	_, err := copySvmResult(res)
	if err != nil {
		return Account{}, err
	}

	return Account{
		Addr:    addr,
		Balance: Amount(account.balance),
		Counter: TxNonce{
			Upper: uint64(account.counter_upper_bits),
			Lower: uint64(account.counter_lower_bits),
		},
	}, nil
}

func (rt *Runtime) CreateAccount(account Account) error {
	res := C.svm_create_genesis_account(
		rt.raw,
		(*C.uchar)(unsafe.Pointer(&account.Addr[0])),
		C.uint64_t(account.Balance),
		C.uint64_t(account.Counter.Upper),
		C.uint64_t(account.Counter.Lower),
	)

	_, err := copySvmResult(res)
	return err
}

// Increases the balance of an Account (i.e printing coins)
//
// # Params
//
// * `addr`   - The `Account Address` we want to increase its balance.
// * `amount` - The `Amount` by which we are going to increase the account's balance.
func (rt *Runtime) IncreaseBalance(addr Address, amount Amount) error {
	res := C.svm_increase_balance(
		rt.raw,
		(*C.uchar)(unsafe.Pointer(&addr[0])),
		C.uint64_t(amount),
	)

	_, err := copySvmResult(res)
	return err
}

func NewEnvelope(principal Address, amount Amount, txNonce TxNonce, gasLimit Gas, gasFee GasFee) *Envelope {
	return &Envelope{
		Principal: principal,
		Amount:    amount,
		TxNonce:   txNonce,
		GasLimit:  gasLimit,
		GasFee:    gasFee,
	}
}

func NewContext(layer Layer, txId TxId) *Context {
	return &Context{
		Layer: layer,
		TxId:  txId,
	}
}

func toSvmParams(env *Envelope, msg []byte, ctx *Context) *svmParams {
	envBytes := encodeEnvelope(env)
	envPtr := (*C.uchar)(unsafe.Pointer(&envBytes[0]))

	msgPtr := (*C.uchar)(unsafe.Pointer(&msg[0]))
	msgLen := (C.uint32_t)(uint32(len(msg)))

	ctxBytes := encodeContext(ctx)
	ctxPtr := (*C.uchar)(unsafe.Pointer(&ctxBytes[0]))

	return &svmParams{
		envPtr,
		msgPtr,
		msgLen,
		ctxPtr,
	}
}

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

	C.svm_free_result(res)
	return receipt, err
}

type svmAction func(params *svmParams) C.svm_result_t
type svmValidation func(rawMsg *C.uchar, msgLen C.uint32_t) C.svm_result_t

func runAction(env *Envelope, msg []byte, ctx *Context, action svmAction) (interface{}, error) {
	if len(msg) == 0 {
		return false, errors.New("`msg` cannot be empty")
	}

	params := toSvmParams(env, msg, ctx)
	res := action(params)
	bytes, err := copySvmResult(res)

	if err != nil {
		return nil, err
	}

	return decodeReceipt(bytes)
}

func runValidation(msg []byte, validator svmValidation) (bool, error) {
	if len(msg) == 0 {
		return false, errors.New("`msg` cannot be empty")
	}

	rawMsg := (*C.uchar)(unsafe.Pointer(&msg[0]))
	msgLen := (C.uint32_t)(len(msg))

	res := validator(rawMsg, msgLen)
	_, err := copySvmResult(res)

	return err == nil, err
}
