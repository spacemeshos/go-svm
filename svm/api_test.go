package svm

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ReadTemplate(t *testing.T, path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	return bytes
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

func TestValidateValidDeploy(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	msg := ReadTemplate(t, "inputs/deploy.svm")
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

	msg := ReadTemplate(t, "inputs/deploy.svm")
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(10), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Deploy(env, msg, ctx)
	assert.Nil(t, err)

	assert.Equal(t, false, receipt.Success)
	assert.Equal(t, receipt.Error.Kind, RuntimeErrorKind(OOG))
}

func DeployWithInfiniteGas(t *testing.T, path string) (*Runtime, *DeployReceipt) {
	Init(true, "")

	rt, _ := NewRuntime()
	defer rt.Destroy()

	msg := ReadTemplate(t, path)
	gas := 1000000000
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(gas), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Deploy(env, msg, ctx)
	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)
	
	return rt, receipt
}

func TestDeploySuccess(t *testing.T) {
	rt, _ := DeployWithInfiniteGas(t, "inputs/deploy.svm")
	rt.Destroy()
}

func TestSpawnSuccess(t *testing.T) {
	rt, _ := DeployWithInfiniteGas(t, "inputs/deploy.svm")
	defer rt.Destroy()

	gas := 1000000000
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(gas), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	msg := []byte{0}
	receipt, err := rt.Spawn(env, msg, ctx)
	assert.Nil(t, err)
	assert.Equal(t, true, receipt.Success)
}