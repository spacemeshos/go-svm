package svm

import (
	"encoding/binary"
)

// * One byte for `tx type`
// * Two bytes for `version`
// * One byte for `success`
const ReceiptHeaderLength = 1 + 2 + 1

func decodeReceipt(bytes []byte) (interface{}, error) {
	return nil, nil
}

func decodeDeployReceipt(bytes []byte) (*DeployReceipt, error) {
	receiptType, success := decodeReceiptHeader(bytes)

	if receiptType != DeployType {
		panic("Expected a `Deploy Receipt`!")
	}

	receipt := &DeployReceipt{
		Success: success,
	}

	return receipt, nil
}

func decodeSpawnReceipt(bytes []byte) (*SpawnReceipt, error) {
	panic("TODO")
}

func decodeCallReceipt(bytes []byte) (*CallReceipt, error) {
	panic("TODO")
}

func decodeError(bytes []byte) error {
	return nil
}

func decodeReceiptHeader(bytes []byte) (TxType, bool) {
	if len(bytes) < ReceiptHeaderLength {
		panic("Received a corrupted Receipt")
	}

	txType := TxType(bytes[0])
	version := int(binary.BigEndian.Uint16(bytes[1:]))
	success := bytes[3] != 0

	if version != 0 {
		panic("For now `version` must be zero")
	}

	return txType, success
}
