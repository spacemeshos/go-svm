package svm

import (
	"encoding/binary"
	"log"
)

// * One byte for `tx type`
// * Two bytes for `version`
// * One byte for `success`
const ReceiptHeaderLength = 1 + 2 + 1

func decodeReceipt(bytes []byte) (interface{}, error) {
	txType, success, bytes := decodeReceiptHeader(bytes)

	if success {
		return decodeSuccess(txType, bytes)
	}
	return decodeFailure(txType, bytes)
}

func decodeFailure(txType TxType, bytes []byte) (interface{}, error) {
	rtError, logs, err := decodeRuntimeError(bytes)
	if err != nil {
		return nil, err
	}

	switch txType {
	case TxType(DeployType):
		receipt := &DeployReceipt{Success: false, Error: rtError, Logs: logs}
		return receipt, nil
	case TxType(SpawnType):
		receipt := &SpawnReceipt{Success: false, Error: rtError, Logs: logs}
		return receipt, nil
	case TxType(CallType):
		receipt := &CallReceipt{Success: false, Error: rtError, Logs: logs}
		return receipt, nil
	default:
		panic("Unreachable")
	}
}

func decodeSuccess(txType TxType, bytes []byte) (interface{}, error) {
	switch txType {
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

func decodeRuntimeError(bytes []byte) (*RuntimeError, []Log, error) {
	errorCode, bytes := decodeErrorCode(bytes)
	logs, bytes := decodeLogs(bytes)
	rtError := &RuntimeError{Kind: errorCode}

	switch errorCode {
	case RuntimeErrorKind(OOG):
		log.Print("OOG")
		return rtError, logs, nil
	case RuntimeErrorKind(TemplateNotFound):
		log.Print("Template Not Found")
		template, _ := decodeAddress(bytes)
		rtError.Template = template
		return rtError, logs, nil
	case RuntimeErrorKind(AccountNotFound):
		log.Print("Account Not Found")
		target, _ := decodeAddress(bytes)
		rtError.Target = target
		return rtError, logs, nil
	case RuntimeErrorKind(CompilationFailed), RuntimeErrorKind(InstantiationFailed):
		log.Print("...")
		template, bytes := decodeAddress(bytes)
		target, bytes := decodeAddress(bytes)
		msg, _ := decodeString(bytes)
		rtError.Template = template
		rtError.Target = target
		rtError.Message = msg
		return rtError, logs, nil
	case RuntimeErrorKind(FuncNotFound):
		template, bytes := decodeAddress(bytes)
		target, bytes := decodeAddress(bytes)
		function, _ := decodeString(bytes)
		rtError.Template = template
		rtError.Target = target
		rtError.Function = function
		return rtError, logs, nil
	case RuntimeErrorKind(FuncNotCtor):
		template, bytes := decodeAddress(bytes)
		function, bytes := decodeString(bytes)
		rtError.Template = template
		rtError.Function = function
		return rtError, logs, nil
	case RuntimeErrorKind(FuncFailed), RuntimeErrorKind(FuncNotAllowed):
		template, bytes := decodeAddress(bytes)
		target, bytes := decodeAddress(bytes)
		function, bytes := decodeString(bytes)
		msg, _ := decodeString(bytes)
		rtError.Template = template
		rtError.Target = target
		rtError.Function = function
		rtError.Message = msg
		return rtError, logs, nil
	case RuntimeErrorKind(FuncInvalidSignature):
		template, bytes := decodeAddress(bytes)
		target, bytes := decodeAddress(bytes)
		function, _ := decodeString(bytes)
		rtError.Template = template
		rtError.Target = target
		rtError.Function = function
		return rtError, logs, nil
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
	returndata, bytes := decodeReturnData(bytes)
	gas, bytes := decodeGas(bytes)
	touchedAccounts, bytes := decodeTouchedAccounts(bytes)
	logs, _ := decodeLogs(bytes)

	receipt := &SpawnReceipt{
		Success:         true,
		AccountAddr:     accountAddr,
		InitState:       initState,
		ReturnData:      returndata,
		GasUsed:         gas,
		TouchedAccounts: touchedAccounts,
		Logs:            logs,
	}
	return receipt, nil
}

func decodeCallReceipt(bytes []byte) (*CallReceipt, error) {
	newState, bytes := decodeState(bytes)
	returndata, bytes := decodeReturnData(bytes)
	gas, bytes := decodeGas(bytes)
	touchedAccounts, bytes := decodeTouchedAccounts(bytes)
	logs, _ := decodeLogs(bytes)

	receipt := &CallReceipt{
		Success:         true,
		NewState:        newState,
		ReturnData:      returndata,
		GasUsed:         gas,
		TouchedAccounts: touchedAccounts,
		Logs:            logs,
	}
	return receipt, nil
}

func decodeTouchedAccounts(bytes []byte) ([]Address, []byte) {
	accounts := []Address{}
	len := int(binary.BigEndian.Uint16(bytes))
	bytes = bytes[2:]
	for i := 0; i < len; i++ {
		addr, newBytes := decodeAddress(bytes)
		bytes = newBytes
		accounts = append(accounts, addr)
	}
	return accounts, bytes
}

func decodeErrorCode(bytes []byte) (RuntimeErrorKind, []byte) {
	return RuntimeErrorKind(bytes[0]), bytes[1:]
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

	return returns, bytes[nextOffset:]
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

func decodeString(bytes []byte) (string, []byte) {
	length := bytes[0]
	data := make([]byte, length)
	nextOffset := 1 + length
	copy(data, bytes[1:nextOffset])

	return string(data), bytes[nextOffset:]
}
