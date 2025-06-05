package health

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/hex"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/encrypt"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Custom error types
var (
	ErrKeeperNotVerified = errors.New("keeper not verified")
)

// ErrorResponse represents the error response from the health service
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Client represents a Health service client
type Client struct {
	httpClient *http.Client
	logger     logging.Logger
	config     Config
}

// Config holds the configuration for the Health client
type Config struct {
	HealthServiceURL string
	PrivateKey       string
	KeeperAddress    string
	PeerID           string
	Version          string
	RequestTimeout   time.Duration
}

// NewClient creates a new Health service client
func NewClient(logger logging.Logger, cfg Config) (*Client, error) {
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10 * time.Second
	}

	if cfg.Version == "" {
		cfg.Version = "0.1.2"
	}

	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
	}

	return &Client{
		httpClient: httpClient,
		logger:     logger,
		config:     cfg,
	}, nil
}

// CheckIn performs a health check-in with the health service
func (c *Client) CheckIn(ctx context.Context) (types.KeeperHealthCheckInResponse, error) {
	// Get consensus address from private key
	privateKey, err := ethcrypto.HexToECDSA(c.config.PrivateKey)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("invalid private key: %w", err)
	}
	publicKeyBytes := ethcrypto.FromECDSAPub(&privateKey.PublicKey)
	consensusPubKey := hex.EncodeToString(publicKeyBytes)
	consensusAddress := ethcrypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Create message to sign
	msg := []byte(c.config.KeeperAddress)
	signature, err := encrypt.SignMessage(string(msg), c.config.PrivateKey)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to sign check-in message: %w", err)
	}

	// Prepare health check payload
	payload := types.KeeperHealthCheckIn{
		KeeperAddress:    c.config.KeeperAddress,
		ConsensusPubKey:  consensusPubKey,
		ConsensusAddress: consensusAddress,
		Version:          c.config.Version,
		Timestamp:        time.Now().UTC(),
		Signature:        signature,
		PeerID:           c.config.PeerID,
	}

	// c.logger.Infof("Payload: %+v", payload)

	// Send health check request
	response, err := c.sendHealthCheck(ctx, payload)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("health check failed: %w", err)
	}

	c.logger.Info("Successfully completed health check-in",
		"status", response.Status,
		"keeperAddress", c.config.KeeperAddress,
		"timestamp", payload.Timestamp)

	return response, nil
}

// sendHealthCheck sends the health check request to the health service
func (c *Client) sendHealthCheck(ctx context.Context, payload types.KeeperHealthCheckIn) (types.KeeperHealthCheckInResponse, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to marshal health check payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/health", c.config.HealthServiceURL),
		bytes.NewBuffer(payloadBytes))
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to send health check request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			if errResp.Code == "KEEPER_NOT_VERIFIED" {
				return types.KeeperHealthCheckInResponse{
					Status: false,
					Data:   errResp.Error,
				}, ErrKeeperNotVerified
			}
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   errResp.Error,
		}, fmt.Errorf("health service returned non-OK status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response types.KeeperHealthCheckInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to unmarshal health check response: %w", err)
	}

	privateKey, err := ethcrypto.HexToECDSA(config.GetPrivateKeyConsensus())
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("invalid private key: %w", err)
	}
	decryptedData, err := DecryptMessageWithKeeperKey(privateKey, response.Data)
	if err != nil {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to decrypt health check response: %w", err)
	}

	parts := bytes.SplitN([]byte(decryptedData), []byte(":"), 2)
	if len(parts) != 2 {
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   "invalid response format",
		}, fmt.Errorf("invalid response format: expected host:token")
	}

	config.SetIpfsHost(string(parts[0]))
	config.SetPinataJWT(string(parts[1]))

	c.logger.Infof("IPFS Host: %s", string(parts[0]))
	c.logger.Infof("Pinata JWT: %s", string(parts[1]))

	return types.KeeperHealthCheckInResponse{
		Status: true,
		Data:   "Health check-in successful",
	}, nil
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

func DecryptMessageWithKeeperKey(privateKey *ecdsa.PrivateKey, encryptedHex string) (string, error) {
	encryptedBytes, err := hexutil.Decode(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("invalid encrypted hex: %w", err)
	}
	eciesPrivKey := ecies.ImportECDSA(privateKey)
	decrypted, err := eciesPrivKey.Decrypt(encryptedBytes, nil, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(decrypted), nil
}