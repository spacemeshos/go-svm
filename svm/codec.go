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

	// `Amount`
	p := AddressLength
	binary.BigEndian.PutUint64(bytes[p:p+AmountLength], (uint64)(env.Amount))

	// `Tx Nonce`
	p += TxNonceLength
	binary.BigEndian.PutUint64(bytes[p:p+TxNonceLength/2], (uint64)(env.TxNonce.Upper))
	binary.BigEndian.PutUint64(bytes[p:p+TxNonceLength/2], (uint64)(env.TxNonce.Lower))

	// `Gas Limit`
	p += GasLength
	binary.BigEndian.PutUint64(bytes[p:p+GasLength], (uint64)(env.GasLimit))

	// `Gas Fee`
	p += GasFeeLength
	binary.BigEndian.PutUint64(bytes[p:p+GasFeeLength], (uint64)(env.GasFee))

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
