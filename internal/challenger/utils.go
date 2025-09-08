package challenger

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	taskResponseType, _ = abi.NewType("tuple", "struct", []abi.ArgumentMarshaling{
		{Name: "TaskID", Type: "uint64"},
		{Name: "IsValid", Type: "bool"},
	})

	Args = abi.Arguments{
		{Type: taskResponseType, Name: "TaskResponse"},
	}
)

type TaskResponse struct {
	TaskID  uint64
	IsValid bool
}

func mustNewType(typeStr string) abi.Type {
	t, err := abi.NewType(typeStr, "", nil)
	if err != nil {
		panic(err)
	}
	return t
}

// MarshalTaskResponse marshals the TaskResponse struct into JSON bytes.
func MarshalTaskResponse(h TaskResponse) ([]byte, error) {
	return json.Marshal(h)
}

// UnmarshalTaskResponse unmarshals the JSON bytes into a TaskResponse struct.
func UnmarshalTaskResponse(jsonData []byte) (TaskResponse, error) {
	var taskResponse TaskResponse
	err := json.Unmarshal(jsonData, &taskResponse)
	return taskResponse, err
}

// GetTaskResponseDigestEncodeByjson returns the hash of the TaskResponse, which is what operators sign over.
func GetTaskResponseDigestEncodeByjson(h TaskResponse) ([32]byte, []byte, error) {
	jsonData, err := MarshalTaskResponse(h)
	if err != nil {
		return [32]byte{}, []byte{}, err
	}
	taskResponseDigest := crypto.Keccak256Hash(jsonData)
	return taskResponseDigest, jsonData, nil
}

// GetTaskResponseDigestEncodeByAbi returns the hash of the TaskResponse using ABI encoding
func GetTaskResponseDigestEncodeByAbi(h TaskResponse) ([32]byte, []byte, error) {
	packed, err := Args.Pack(h)
	if err != nil {
		return [32]byte{}, []byte{}, err
	}
	taskResponseDigest := crypto.Keccak256Hash(packed)
	return taskResponseDigest, packed, nil
}

// AbiEncode encodes a TaskResponse using ABI encoding (for compatibility)
func AbiEncode(resp TaskResponse) ([]byte, error) {
	abiArgs := abi.Arguments{
		{Name: "taskID", Type: mustNewType("uint64")},
		{Name: "isValid", Type: mustNewType("bool")},
	}

	taskID := new(big.Int).SetUint64(resp.TaskID)
	isValid := resp.IsValid

	return abiArgs.Pack(taskID, isValid)
}

// AbiDecode decodes ABI-encoded data into a TaskResponse (for compatibility)
func AbiDecode(data []byte) (TaskResponse, error) {
	abiArgs := abi.Arguments{
		{Name: "taskID", Type: mustNewType("uint64")},
		{Name: "isValid", Type: mustNewType("bool")},
	}

	values, err := abiArgs.UnpackValues(data)
	if err != nil {
		return TaskResponse{}, err
	}

	var (
		taskID, _  = values[0].(*big.Int)
		isValid, _ = values[1].(bool)
	)

	return TaskResponse{
		TaskID:  taskID.Uint64(),
		IsValid: isValid,
	}, nil
}

// ===== REMOVED KEEPER-SPECIFIC FUNCTIONS =====
// The following functions were removed as they are not required for challenger functionality:
// - UpdateYAMLWithComments (file manipulation not needed)
// - GetFileInCurrentDirectory (file operations not needed)
// - ConvertToEthAddresses (address conversion utilities not needed)
// - SwitchEthAddressToImAddress (imua-specific address conversion not needed)
// - ChainIDWithoutRevision (chain ID manipulation not needed)
// - BLSMessageToSign constant (not used in challenger)
// - IsRevisionFormat regex (not used in challenger)