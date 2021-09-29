package svm

import (
	"encoding/binary"
)

// * One byte for `tx type`
// * Two bytes for `version`
// * One byte for `success`
const ReceiptHeaderLength = 1 + 2 + 1

func decodeReceipt(bytes []byte) (interface{}, error) {
	receiptType, success, bytes := decodeReceiptHeader(bytes)

	if !success {
		panic("TODO: parse error")
	}

	switch receiptType {
	case TxType(DeployType):
		return decodeDeployReceipt(bytes)
	case TxType(SpawnType):
		return decodeSpawnReceipt(bytes)
	case TxType(CallType):
		return decodeCallReceipt(bytes)
	default:
		panic("Unreachable")
	}
}

func decodeDeployReceipt(bytes []byte) (*DeployReceipt, error) {
	templateAddr, bytes := decodeAddress(bytes)
	gas, bytes := decodeGas(bytes)
	logs, _ := decodeLogs(bytes)

	receipt := &DeployReceipt{
		Success:      true,
		TemplateAddr: templateAddr,
		GasUsed:      gas,
		Logs:         logs,
	}
	return receipt, nil
}

func decodeSpawnReceipt(bytes []byte) (*SpawnReceipt, error) {
	accountAddr, bytes := decodeAddress(bytes)
	initState, bytes := decodeState(bytes)
	gas, bytes := decodeGas(bytes)
	returndata, bytes := decodeReturnData(bytes)
	logs, _ := decodeLogs(bytes)

	receipt := &SpawnReceipt{
		Success:     true,
		AccountAddr: accountAddr,
		InitState:   initState,
		ReturnData:  returndata,
		GasUsed:     gas,
		Logs:        logs,
	}
	return receipt, nil
}

func decodeCallReceipt(bytes []byte) (*CallReceipt, error) {
	newState, bytes := decodeState(bytes)
	gas, bytes := decodeGas(bytes)
	returndata, bytes := decodeReturnData(bytes)
	logs, _ := decodeLogs(bytes)

	receipt := &CallReceipt{
		Success:    true,
		NewState:   newState,
		ReturnData: returndata,
		GasUsed:    gas,
		Logs:       logs,
	}
	return receipt, nil
}

func decodeError(bytes []byte) error {
	panic("TODO")
}

func decodeReceiptHeader(bytes []byte) (TxType, bool, []byte) {
	if len(bytes) < ReceiptHeaderLength {
		panic("Received a corrupted Receipt")
	}

	txType := TxType(bytes[0])
	version := int(binary.BigEndian.Uint16(bytes[1:]))
	success := bytes[3] != 0

	if version != 0 {
		panic("For now `version` must be zero")
	}

	return txType, success, bytes[ReceiptHeaderLength:]
}

func decodeReturnData(bytes []byte) (ReturnData, []byte) {
	returnsSize := int(binary.BigEndian.Uint16(bytes))
	offset := 2

	returns := make([]byte, returnsSize)
	nextOffset := offset + returnsSize
	copy(returns, bytes[offset:nextOffset])

	return nil, bytes[nextOffset:]
}

func decodeLogs(bytes []byte) ([]Log, []byte) {
	logsCount := bytes[0]
	logs := make([]Log, logsCount)

	offset := 1
	for i := 0; i < int(logsCount); i++ {
		logLength := int(binary.BigEndian.Uint16(bytes[offset:]))
		offset += 2
		log := make([]byte, logLength)
		nextOffset := offset + logLength
		copy(log, bytes[offset:nextOffset])

		logs[i] = log
		offset = nextOffset
	}

	return logs, bytes[offset:]
}
