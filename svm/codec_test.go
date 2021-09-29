package svm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeEnvelope(bytes [EnvelopeLength]byte) *Envelope {
	return nil
}

func TestEncodeDecodeEnvelope(t *testing.T) {
	env := NewEnvelope(Address{0}, Amount(0), TxNonce{Upper: 0, Lower: 0}, Gas(0), GasFee(0))
	bytes := encodeEnvelope(env)

	assert.Equal(t, env, decodeEnvelope(bytes))
}
