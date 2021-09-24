package svm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	err := Init(true, "")
	assert.Nil(t, err)
}
