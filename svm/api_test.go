package svm

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFile(t *testing.T, path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	return bytes
}

func runtimeSetup(t *testing.T) *Runtime {
	api, err := Init()
	assert.Nil(t, err)

	rt, err := api.NewRuntime(true, "")
	assert.NotNil(t, rt)
	assert.Nil(t, err)

	return rt
}

type TestParams struct {
	Amount    Amount
	Principal Address
	Nonce     TxNonce
	TxId      TxId
	Gas       Gas
	GasFee    GasFee
	Layer     Layer
}

func NewTestParams() *TestParams {
	// # Note
	// We inject by default a super high Gas to avoid running Out-of-Gas
	return &TestParams{
		Amount:    Amount(0),
		Principal: Address{},
		Nonce:     TxNonce{Upper: 0, Lower: 0},
		TxId:      TxId{},
		Gas:       Gas(1000000000),
		GasFee:    GasFee(0),
		Layer:     Layer(0),
	}
}

func executeTx(t *testing.T, rt *Runtime, path string, params *TestParams, f func(*Runtime, *Envelope, []byte, *Context) (interface{}, error)) (interface{}, error) {
	msg := readFile(t, path)
	env := NewEnvelope(params.Principal, params.Amount, params.Nonce, params.Gas, params.GasFee)
	ctx := NewContext(params.Layer, params.TxId)

	receipt, err := f(rt, env, msg, ctx)
	return receipt, err
}

func deploy(t *testing.T, rt *Runtime, path string, params *TestParams) (*DeployReceipt, error) {
	receipt, err :=
		executeTx(t, rt, path, params, func(rt *Runtime, env *Envelope, msg []byte, ctx *Context) (interface{}, error) {
			return rt.Deploy(env, msg, ctx)
		})

	return receipt.(*DeployReceipt), err
}

func spawn(t *testing.T, rt *Runtime, path string, params *TestParams) (*SpawnReceipt, error) {
	receipt, err :=
		executeTx(t, rt, path, params, func(rt *Runtime, env *Envelope, msg []byte, ctx *Context) (interface{}, error) {
			return rt.Spawn(env, msg, ctx)
		})

	return receipt.(*SpawnReceipt), err
}

func call(t *testing.T, rt *Runtime, path string, params *TestParams) (*CallReceipt, error) {
	receipt, err :=
		executeTx(t, rt, path, params, func(rt *Runtime, env *Envelope, msg []byte, ctx *Context) (interface{}, error) {
			return rt.Call(env, msg, ctx)
		})

	return receipt.(*CallReceipt), err
}

func TestInitNilErr(t *testing.T) {
	_, err := Init()
	assert.Nil(t, err)
}

func TestNewRuntime(t *testing.T) {
	api, err := Init()
	assert.Nil(t, err)

	assert.Equal(t, 0, api.RuntimesCount())
	rt := runtimeSetup(t)

	assert.Equal(t, 1, api.RuntimesCount())
	rt.Destroy()
	assert.Equal(t, 0, api.RuntimesCount())
}

func TestValidateEmptyDeploy(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	ok, err := rt.ValidateDeploy([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestValidateDeployInvalid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := []byte{0, 0, 0, 0}
	valid, err := rt.ValidateDeploy(msg)
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestValidateDeployValid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := readFile(t, "inputs/template_example.svm")
	valid, err := rt.ValidateDeploy(msg)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestValidateEmptySpawn(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	ok, err := rt.ValidateSpawn([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestValidateEmptyCall(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	ok, err := rt.ValidateCall([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestDeployOutOfGas(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := readFile(t, "inputs/template_example.svm")
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(10), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Deploy(env, msg, ctx)
	assert.Nil(t, err)

	assert.Equal(t, false, receipt.Success)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(OOG))
}

func TestDeploySuccess(t *testing.T) {
	rt := runtimeSetup(t)
	receipt, err := deploy(t, rt, "inputs/template_example.svm", NewTestParams())
	defer rt.Destroy()

	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)
}

func TestSpawnValidateInvalid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := []byte{0, 0, 0, 0}
	isValid, _ := rt.ValidateSpawn(msg)
	assert.False(t, isValid)
}

func TestSpawnValidateValid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := readFile(t, "inputs/spawn/initialize.json.bin")
	isValid, _ := rt.ValidateSpawn(msg)
	assert.True(t, isValid)
}

func TestSpawnOutOfGas(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())

	params := NewTestParams()
	params.Gas = Gas(10)
	receipt, err := spawn(t, rt, "inputs/spawn/initialize.json.bin", params)

	assert.Nil(t, err)
	assert.Equal(t, false, receipt.Success)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(OOG))
}

func TestSpawnSuccess(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())
	receipt, err := spawn(t, rt, "inputs/spawn/initialize.json.bin", NewTestParams())

	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)
	assert.NotNil(t, receipt.InitState)
	assert.NotNil(t, receipt.AccountAddr)

	assert.Len(t, receipt.TouchedAccounts, 2)
	assert.Contains(t, receipt.TouchedAccounts, receipt.AccountAddr)
	assert.Contains(t, receipt.TouchedAccounts, Address{})
}

func TestCallValidateInvalid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := []byte{0, 0, 0, 0}
	isValid, _ := rt.ValidateCall(msg)
	assert.False(t, isValid)
}

func TestCallValidateValid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	msg := readFile(t, "inputs/call/load_addr.json.bin")
	isValid, _ := rt.ValidateCall(msg)
	assert.True(t, isValid)
}

func TestAccountBalance(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	err := rt.CreateAccount(Account{Addr: Address{}, Balance: Amount(10), Counter: TxNonce{Upper: 0, Lower: 1337}})
	assert.Nil(t, err)

	account, err := rt.GetAccount(Address{})
	assert.Nil(t, err)
	assert.Equal(t, account.Addr, Address{})
	assert.Equal(t, account.Balance, Amount(10))
	assert.Equal(t, account.Counter, TxNonce{Upper: 0, Lower: 1337})

	err = rt.IncreaseBalance(Address{}, Amount(10))
	assert.Nil(t, err)

	account, err = rt.GetAccount(Address{})
	assert.Nil(t, err)
	assert.Equal(t, account.Balance, Amount(20))
}

// TODO: fix the gas pricing first under SVM
// func TestCallOutOfGas(t *testing.T) {
// 	rt := runtimeSetup(t)
// 	defer rt.Destroy()

// 	deploy(t, rt, "inputs/template_example.svm")
// 	spawn(t, rt, "inputs/spawn/initialize.json.bin")

// 	msg := readFile(t, "inputs/call/store_addr.json.bin")
// 	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(10), GasFee(0))
// 	ctx := NewContext(Layer(0), TxId{})

// 	receipt, err := rt.Call(env, msg, ctx)
// 	assert.Nil(t, err)

// 	assert.Equal(t, false, receipt.Success)
// 	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(OOG))
// }

func TestSpawnNonexistentCtor(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())

	receipt, err := spawn(t, rt, "inputs/spawn/nonexistent_ctor.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.False(t, receipt.Success)
	assert.NotNil(t, receipt.Error)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(FuncNotCtor))
}

func TestSpawnCtorExistsButInvalid(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())

	receipt, err := spawn(t, rt, "inputs/spawn/ctor_exists_but_invalid.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.False(t, receipt.Success)
	assert.NotNil(t, receipt.Error)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(FuncNotCtor))
}

func TestCallNonexistentFunc(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())
	spawn(t, rt, "inputs/spawn/initialize.json.bin", NewTestParams())

	receipt, err := call(t, rt, "inputs/call/nonexistent_func.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.False(t, receipt.Success)
	assert.NotNil(t, receipt.Error)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(FuncNotFound))
}

func TestCallCtor(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())
	spawn(t, rt, "inputs/spawn/initialize.json.bin", NewTestParams())

	receipt, err := call(t, rt, "inputs/call/initialize.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.False(t, receipt.Success)
	assert.NotNil(t, receipt.Error)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(FuncNotAllowed))
}

func TestCallSuccess(t *testing.T) {
	rt := runtimeSetup(t)
	defer rt.Destroy()

	deploy(t, rt, "inputs/template_example.svm", NewTestParams())
	spawn(t, rt, "inputs/spawn/initialize.json.bin", NewTestParams())
	receipt, err := call(t, rt, "inputs/call/store_addr.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)

	targetAddr := Address{0x06, 0x68, 0x18, 0xab, 0xe3, 0x61, 0xdd, 0x44, 0xf4, 0x25, 0xda, 0x19, 0xe1, 0x7c, 0x45, 0xba, 0xbc, 0x40, 0xe2, 0x32}
	assert.Len(t, receipt.TouchedAccounts, 1)
	assert.Contains(t, receipt.TouchedAccounts, targetAddr)

	receipt, err = call(t, rt, "inputs/call/load_addr.json.bin", NewTestParams())
	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)

	returns := receipt.ReturnData
	assert.Equal(t, len(returns), 1+AddressLength)

	// type is `Address`
	assert.Equal(t, returns[0], byte(0x40))

	// expected loaded Address to be `102030405060708090102030405060708090AABB`
	assert.Equal(t, returns[1:], []byte{16, 32, 48, 64, 80, 96, 112, 128, 144, 16, 32, 48, 64, 80, 96, 112, 128, 144, 170, 187})
}
