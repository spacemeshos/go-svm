package svm

import (
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
	ok, err := rt.ValidateDeploy([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	rt.Destroy()
}

func TestValidateEmptySpawn(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	ok, err := rt.ValidateSpawn([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	rt.Destroy()
}

func TestValidateEmptyCall(t *testing.T) {
	Init(true, "")

	rt, _ := NewRuntime()
	ok, err := rt.ValidateCall([]byte{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	rt.Destroy()
}
