package svm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func decodeEnvelope(bytes []byte) (*Envelope, []byte) {
	principal, bytes := decodeAddress(bytes)
	amount, bytes := decodeAmount(bytes)
	txNonce, bytes := decodeTxNonce(bytes)
	gasLimit, bytes := decodeGas(bytes)
	gasFee, bytes := decodeGasFeeUsed(bytes)
	env := NewEnvelope(principal, amount, txNonce, gasLimit, gasFee)

	return env, bytes
}

func decodeContext(bytes []byte) (*Context, []byte) {
	layer, bytes := decodeLayer(bytes)
	txId, bytes := decodeTxId(bytes)
	ctx := NewContext(layer, txId)

	return ctx, bytes
}

func TestEncodeDecodeEnvelope(t *testing.T) {
	expected := NewEnvelope(Address{0xFF}, Amount(100), TxNonce{Upper: 0xAB, Lower: 0xCD}, Gas(200), GasFee(5))
	bytes := encodeEnvelope(expected)
	actual, _ := decodeEnvelope(bytes[:])

	assert.Equal(t, expected, actual)
}

func TestEncodeDecodeContext(t *testing.T) {
	expected := NewContext(Layer(10), TxId{0xFF})
	bytes := encodeContext(expected)
	actual, _ := decodeContext(bytes[:])

	assert.Equal(t, expected, actual)
}
