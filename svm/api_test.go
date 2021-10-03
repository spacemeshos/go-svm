package svm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func ReadTemplate(t *testing.T, path string) []byte {
	bytes, err := os.ReadFile(path)
	assert.Nil(t, err)
	return bytes
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
	env := NewEnvelope(Address{}, Amount(10), TxNonce{Upper: 0, Lower: 0}, Gas(1000000), GasFee(0))
	ctx := NewContext(Layer(0), TxId{})

	receipt, err := rt.Deploy(env, msg, ctx)
	assert.Nil(t, err)

	t.Log(receipt)
	// assert.Equal(t, false, receipt.Success)
}
