package checkin

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

func CheckInWithHealthService() error {
	healthServiceURL := fmt.Sprintf("%s/health", config.HealthRPCAddress)

	privateKeyHex := config.PrivateKeyConsensus
	operatorAddress := config.KeeperAddress

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		logger.Error("Invalid private key", "error", err)
		return fmt.Errorf("invalid private key: %w", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	consensusAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	msg := []byte(operatorAddress)
	msgHash := crypto.Keccak256Hash(msg)
	signatureBytes, err := crypto.Sign(msgHash.Bytes(), privateKey)
	if err != nil {
		logger.Error("Failed to sign check-in message", "error", err)
		return fmt.Errorf("failed to sign check-in message: %w", err)
	}
	signature := "0x" + common.Bytes2Hex(signatureBytes)

	payload := types.KeeperHealth{
		KeeperAddress:    operatorAddress,
		ConsensusAddress: consensusAddress,
		Version:          "0.1.0",
		Timestamp:        time.Now().UTC(),
		Signature:        signature,
		PeerID:           config.PeerID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal check-in payload", "error", err)
		return fmt.Errorf("failed to marshal check-in payload: %w", err)
	}

	req, err := http.NewRequest("POST", healthServiceURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to create check-in request", "error", err)
		return fmt.Errorf("failed to create check-in request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send check-in request", "error", err)
		return fmt.Errorf("failed to send check-in request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Health service returned non-OK status", "status", resp.StatusCode)
		return fmt.Errorf("health service returned status: %d", resp.StatusCode)
	}

	return nil
}
