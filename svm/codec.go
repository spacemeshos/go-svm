package svm

import "encoding/binary"

// In other words, holds fields which are part of any transaction regardless of its type (i.e `Deploy/Spawn/Call`).
/// Encoding of a binary [`Envelope`].
///
/// ```text
///  +-------------+--------------+----------------+----------------+----------------+
///  |             |              |			       |                |				 |
///  |  Principal  |    Amount    |    Tx Nonce    |   Gas Limit    |    Gas Fee 	 |
///  |  (Address)  |    (u64)     |     (u128)     |     (u64)      |     (u64)	 	 |
///  |             |              |                |                |				 |
///  |  20 bytes   |   8 bytes    |    16 bytes    |    8 bytes     |    8 bytes     |
///  |             | (Big-Endian) |  (Big-Endian)  |  (Big-Endian)  |  (Big-Endian)  |
///  |             |              |                |                |			     |
///  +-------------+--------------+----------------+----------------+----------------+
/// ```
func encodeEnvelope(env *Envelope) [EnvelopeLength]byte {
	bytes := [EnvelopeLength]byte{0}

	// `Principal`
	copy(bytes[:AddressLength], env.Principal[:])
	off := AddressLength

	// `Amount`
	binary.BigEndian.PutUint64(bytes[off:off+AmountLength], (uint64)(env.Amount))
	off += AmountLength

	// `Tx Nonce`
	binary.BigEndian.PutUint64(bytes[off:off+TxNonceLength/2], (uint64)(env.TxNonce.Upper))
	off += TxNonceLength/2
	binary.BigEndian.PutUint64(bytes[off:off+TxNonceLength/2], (uint64)(env.TxNonce.Lower))
	off += TxNonceLength/2

	// `Gas Limit`
	binary.BigEndian.PutUint64(bytes[off:off+GasLength], (uint64)(env.GasLimit))
	off += GasLength

	// `Gas Fee`
	binary.BigEndian.PutUint64(bytes[off:off+GasFeeLength], (uint64)(env.GasFee))
	off += GasFeeLength

	return bytes
}

/// ```text
///  +-------------+--------------+
///  |             |              |
///  |    Layer    |    Tx Id     |
///  |   (u64)     |    (Blob)    |
///  |             |              |
///  |  8 bytes    |   32 bytes   |
///  |             | 			  |
///  +-------------+--------------+
/// ```
func encodeContext(ctx *Context) [ContextLength]byte {
	bytes := [ContextLength]byte{0}

	// `Layer`
	p := 0
	binary.BigEndian.PutUint64(bytes[:LayerLength], (uint64)(ctx.Layer))

	// `Tx Id`
	p += LayerLength
	copy(bytes[p:p+TxIdLength], ctx.TxId[:])

	return bytes
}

func decodeAddress(bytes []byte) ([AddressLength]byte, []byte) {
	var addr [AddressLength]byte
	copy(addr[:], bytes[:AddressLength])

	return addr, bytes[AddressLength:]
}

func decodeState(bytes []byte) ([StateLength]byte, []byte) {
	var state [StateLength]byte
	copy(state[:], bytes[:StateLength])

	return state, bytes[StateLength:]
}

func decodeTxId(bytes []byte) (TxId, []byte) {
	var txId [TxIdLength]byte
	copy(txId[:], bytes[:TxIdLength])

	return TxId(txId), bytes[TxIdLength:]
}

func decodeGas(bytes []byte) (Gas, []byte) {
	gas := binary.BigEndian.Uint64(bytes)
	return Gas(gas), bytes[GasLength:]
}

func decodeGasFeeUsed(bytes []byte) (GasFee, []byte) {
	gas := binary.BigEndian.Uint64(bytes)
	return GasFee(gas), bytes[GasFeeLength:]
}

func decodeAmount(bytes []byte) (Amount, []byte) {
	amount := binary.BigEndian.Uint64(bytes)
	return Amount(amount), bytes[AmountLength:]
}

func decodeLayer(bytes []byte) (Layer, []byte) {
	layer := binary.BigEndian.Uint64(bytes)
	return Layer(layer), bytes[LayerLength:]
}

func decodeTxNonce(bytes []byte) (TxNonce, []byte) {
	upper := binary.BigEndian.Uint64(bytes)
	lower := binary.BigEndian.Uint64(bytes[8:])

	return TxNonce{Upper: upper, Lower: lower}, bytes[16:]
}
