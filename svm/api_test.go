package svm

import (
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFile(t *testing.T, path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	return bytes
}

func deploy(t *testing.T, path string) (*Runtime, *DeployReceipt, error) {
	Init(true, "")

	rt, err := NewRuntime()
	assert.Nil(t, err)

	msg := readFile(t, path)
	gas := 1000000000
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(gas), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Deploy(env, msg, ctx)
	return rt, receipt, err
}

func spawn(t *testing.T, rt *Runtime, path string) (*SpawnReceipt, error) {
	msg := readFile(t, path)
	gas := 1000000000
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(gas), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Spawn(env, msg, ctx)
	return receipt, err
}

func TestInitMemoryNilErr(t *testing.T) {
	assert.Equal(t, 0, RuntimesCount())
	err := Init(true, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, RuntimesCount())
}

func TestNewRuntime(t *testing.T) {
	assert.Equal(t, 0, RuntimesCount())
	Init(true, "")

	rt, err := NewRuntime()
	assert.NotNil(t, rt)
	assert.Nil(t, err)

	assert.Equal(t, 1, RuntimesCount())
	rt.Destroy()
	assert.Equal(t, 0, RuntimesCount())
}

func TestValidateEmptyDeploy(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	ok, err := rt.ValidateDeploy([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestValidateDeployInvalid(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	msg := []byte{0, 0, 0, 0}
	valid, err := rt.ValidateDeploy(msg)
	assert.False(t, valid)
	assert.NotNil(t, err)
}

func TestValidateDeployValid(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	msg := readFile(t, "inputs/template_example.svm")
	valid, err := rt.ValidateDeploy(msg)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestValidateEmptySpawn(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	ok, err := rt.ValidateSpawn([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestValidateEmptyCall(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	ok, err := rt.ValidateCall([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestDeployOutOfGas(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
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
	rt, receipt, err := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)
}

func TestSpawnValidateInvalid(t *testing.T) {
	rt, _, _ := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	msg := []byte{0, 0, 0, 0}
	isValid, _ := rt.ValidateSpawn(msg)
	assert.False(t, isValid)
}

func TestSpawnValidateValid(t *testing.T) {
	rt, _, _ := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	msg := readFile(t, "inputs/spawn/spawn-1.json.bin")
	isValid, _ := rt.ValidateSpawn(msg)
	assert.True(t, isValid)
}

func TestSpawnOutOfGas(t *testing.T) {
	rt, _, _ := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	msg := readFile(t, "inputs/spawn/spawn-1.json.bin")
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(10), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Spawn(env, msg, ctx)
	assert.Nil(t, err)

	assert.Equal(t, false, receipt.Success)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(OOG))
}

func TestSpawnSuccess(t *testing.T) {
	rt, _, _ := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	receipt, err := spawn(t, rt, "inputs/spawn/initialize.json.bin")
	assert.Nil(t, err)

	assert.Equal(t, true, receipt.Success)
	assert.NotNil(t, receipt.InitState)
	assert.NotNil(t, receipt.AccountAddr)

	t.Log(hex.Dump(receipt.InitState[:]))
}

func TestCallSuccess(t *testing.T) {
	// Deploy
	rt, _, _ := deploy(t, "inputs/template_example.svm")
	defer rt.Destroy()

	// Spawn
	receipt, err := spawn(t, rt, "inputs/spawn/initialize.json.bin")
	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)

	// Call
	receipt, err = spawn(t, rt, "inputs/call/store_addr.json.bin")
	assert.Nil(t, err)

	t.Log(receipt.Error)
	// assert.Equal(t, true, receipt.Success)
	// assert.NotNil(t, receipt.InitState)
	// assert.NotNil(t, receipt.AccountAddr)
}
