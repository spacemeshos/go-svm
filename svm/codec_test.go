package svm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeEnvelope(bytes []byte) *Envelope {
	principal, bytes := decodeAddress(bytes)
	amount, bytes := decodeAmount(bytes)
	txNonce, bytes := decodeTxNonce(bytes)
	gasLimit, bytes := decodeGas(bytes)
	gasFee, bytes := decodeGasFeeUsed(bytes)

	return NewEnvelope(principal, amount, txNonce, gasLimit, gasFee)
}

func decodeContext(bytes [ContextLength]byte) *Context {
	layer, bytes := decodeLayer(bytes)
	txId, bytes := decodeTxI(bytes)
	
}

func TestEncodeDecodeEnvelope(t *testing.T) {
	env := NewEnvelope(Address{0}, Amount(0), TxNonce{Upper: 0, Lower: 0}, Gas(0), GasFee(0))
	bytes := encodeEnvelope(env)

	assert.Equal(t, env, decodeEnvelope(bytes))
}
