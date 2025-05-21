package aggregator

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Common errors
var (
	ErrInvalidKey    = fmt.Errorf("invalid key")
	ErrSigningFailed = fmt.Errorf("signing operation failed")
	ErrRPCFailed     = fmt.Errorf("RPC operation failed")
	ErrMarshalFailed = fmt.Errorf("marshaling operation failed")
)

// AggregatorClientConfig holds the configuration for AggregatorClient
type AggregatorClientConfig struct {
	RPCAddress string
	PrivateKey string
	RPCTimeout time.Duration
}

// taskParams represents the parameters for sending a task
type taskParams struct {
	proofOfTask      string
	data             string
	taskDefinitionID int
	performerAddress string
	signature        string
}

// AggregatorClient handles communication with the aggregator service
type AggregatorClient struct {
	logger     logging.Logger
	config     AggregatorClientConfig
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// NewAggregatorClient creates a new instance of AggregatorClient
func NewAggregatorClient(logger logging.Logger, cfg AggregatorClientConfig) (*AggregatorClient, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.RPCAddress == "" {
		return nil, fmt.Errorf("RPC address cannot be empty")
	}
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}
	if cfg.RPCTimeout <= 0 {
		cfg.RPCTimeout = 10 * time.Second
	}

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to convert private key: %v", ErrInvalidKey, err)
	}

	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: invalid public key type", ErrInvalidKey)
	}

	return &AggregatorClient{
		logger:     logger,
		config:     cfg,
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Close cleans up any resources used by the client
func (c *AggregatorClient) Close() error {
	// Currently no resources to clean up
	return nil
}

// getPerformerAddress returns the performer's Ethereum address
func (c *AggregatorClient) getPerformerAddress() string {
	return crypto.PubkeyToAddress(*c.publicKey).Hex()
}

// signMessage signs the given data with the client's private key
func (c *AggregatorClient) signMessage(data []byte) (string, error) {
	messageHash := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(messageHash.Bytes(), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSigningFailed, err)
	}

	sig[64] += 27
	return hexutil.Encode(sig), nil
}

// SendTaskToPerformer sends a task to the specified performer through the aggregator
func (c *AggregatorClient) SendTaskToPerformer(jobData *types.HandleCreateJobData, triggerData *types.TriggerData, performerData types.GetPerformerData) (bool, error) {
	c.logger.Debug("Sending task to performer",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID)

	// Pack task data
	data := map[string]interface{}{
		"jobData":       jobData,
		"triggerData":   triggerData,
		"performerData": performerData,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.logger.Error("Failed to marshal task data", "error", err)
		return false, fmt.Errorf("%w: %v", ErrMarshalFailed, err)
	}

	// Prepare ABI arguments
	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	performerAddress := c.getPerformerAddress()

	dataPacked, err := arguments.Pack(
		"proofOfTask",
		jsonData,
		common.HexToAddress(performerAddress),
		big.NewInt(0),
	)
	if err != nil {
		c.logger.Error("Failed to encode task data", "error", err)
		return false, fmt.Errorf("%w: failed to encode data: %v", ErrMarshalFailed, err)
	}

	signature, err := c.signMessage(dataPacked)
	if err != nil {
		c.logger.Error("Failed to sign task data", "error", err)
		return false, err
	}

	c.logger.Debug("Task data signed successfully", "signature", signature)

	// Connect to RPC
	client, err := rpc.Dial(c.config.RPCAddress)
	if err != nil {
		c.logger.Error("Failed to connect to RPC", "error", err)
		return false, fmt.Errorf("%w: failed to dial: %v", ErrRPCFailed, err)
	}
	defer client.Close()

	// Prepare and send RPC request
	params := taskParams{
		proofOfTask:      "proofOfTask",
		data:             "0x" + hex.EncodeToString(jsonData),
		taskDefinitionID: 0,
		performerAddress: performerAddress,
		signature:        signature,
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.RPCTimeout)
	defer cancel()

	var result interface{}
	err = client.CallContext(ctx, &result, "sendCustomMessage", params.data, params.taskDefinitionID)
	if err != nil {
		c.logger.Error("RPC request failed", "error", err)
		return false, fmt.Errorf("%w: %v", ErrRPCFailed, err)
	}

	c.logger.Info("Task sent successfully",
		"performerID", performerData.KeeperID,
		"jobID", jobData.JobID,
		"result", result)
	return true, nil
}
