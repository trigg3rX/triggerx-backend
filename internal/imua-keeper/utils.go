package keeper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cosmos/btcutil/bech32"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v3"
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

const BLSMessageToSign = "BLS12-381 Signed Message\nChainIDWithoutRevision: %s\nAccAddressBech32: %s"

func AbiEncode(resp TaskResponse) ([]byte, error) {

	abiArgs := abi.Arguments{
		{Name: "taskID", Type: mustNewType("uint64")},
		{Name: "isValid", Type: mustNewType("bool")},
	}

	taskID := new(big.Int).SetUint64(resp.TaskID)
	isValid := resp.IsValid

	return abiArgs.Pack(taskID, isValid)
}

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

func mustNewType(typeStr string) abi.Type {
	t, err := abi.NewType(typeStr, "", nil)
	if err != nil {
		panic(err)
	}
	return t
}

// MarshalTaskResponse GetTaskResponseDigestEncodeByjson returns the hash of the TaskResponse, which is what operators sign over
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
func GetTaskResponseDigestEncodeByAbi(h TaskResponse) ([32]byte, []byte, error) {
	packed, err := Args.Pack(h)
	if err != nil {
		return [32]byte{}, []byte{}, err
	}
	taskResponseDigest := crypto.Keccak256Hash(packed)
	return taskResponseDigest, packed, nil
}
func UpdateYAMLWithComments(filePath, key, newValue string) error {
	if newValue == "" {
		panic("param is nil")
	}
	// Read the original YAML file content
	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	// Parse YAML using yaml.v3 node parser to preserve comments
	var doc yaml.Node
	err = yaml.Unmarshal(data, &doc)
	if err != nil {
		return err
	}
	// Iterate through YAML content to find and update the specified key
	for i := 0; i < len(doc.Content[0].Content); i += 2 {
		if doc.Content[0].Content[i].Value == key {
			doc.Content[0].Content[i+1].Kind = yaml.ScalarNode
			doc.Content[0].Content[i+1].Value = newValue
			doc.Content[0].Content[i+1].Tag = "tag:yaml.org,2002:str"
			break
		}
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	err = encoder.Encode(&doc)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
func GetFileInCurrentDirectory(filename string) (string, error) {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Construct full file path
	fullPath := filepath.Join(currentDir, filename)

	// Check if file exists
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("file %s not found in current directory", filename)
	}

	return fullPath, nil
}
func ConvertToEthAddresses(strArray []string) []common.Address {
	var ethAddresses []common.Address

	if len(strArray) > 0 {
		for _, str := range strArray {
			address := common.HexToAddress(str)
			ethAddresses = append(ethAddresses, address)
		}
	}

	return ethAddresses
}
func SwitchEthAddressToImAddress(ethAddress string) (string, error) {
	b, err := hex.DecodeString(ethAddress[2:])
	if err != nil {
		return "", fmt.Errorf("failed to decode eth address: %w", err)
	}

	// Generate im address
	bech32Prefix := "im"
	imAddress, err := bech32.EncodeFromBase256(bech32Prefix, b)
	if err != nil {
		return "", fmt.Errorf("failed to encode bech32 address: %w", err)
	}

	return imAddress, nil
}

// ChainIDWithoutRevision returns the chainID without the revision number.
// For example, "imuachaintestnet_233-1" returns "imuachaintestnet_233".
func ChainIDWithoutRevision(chainID string) string {
	if !IsRevisionFormat(chainID) {
		return chainID
	}
	splitStr := strings.Split(chainID, "-")
	return splitStr[0]
}

var IsRevisionFormat = regexp.MustCompile(`^.*[^\n-]-{1}[1-9][0-9]*$`).MatchString
