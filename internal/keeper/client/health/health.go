package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/security"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/types"
)

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
func (c *Client) CheckIn(ctx context.Context) error {
	// Get consensus address from private key
	privateKey, err := ethcrypto.HexToECDSA(c.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}
	consensusAddress := ethcrypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Create message to sign
	msg := []byte(c.config.KeeperAddress)
	signature, err := security.SignMessage(string(msg), c.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign check-in message: %w", err)
	}

	// Prepare health check payload
	payload := types.KeeperHealthCheckIn{
		KeeperAddress:    c.config.KeeperAddress,
		ConsensusAddress: consensusAddress,
		Version:          c.config.Version,
		Timestamp:        time.Now().UTC(),
		Signature:        signature,
		PeerID:           c.config.PeerID,
	}

	c.logger.Infof("Payload: %+v", payload)

	// Send health check request
	err = c.sendHealthCheck(ctx, payload)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	c.logger.Info("Successfully completed health check-in",
		"keeperAddress", c.config.KeeperAddress,
		"timestamp", payload.Timestamp)

	return nil
}

// sendHealthCheck sends the health check request to the health service
func (c *Client) sendHealthCheck(ctx context.Context, payload types.KeeperHealthCheckIn) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal health check payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/health", c.config.HealthServiceURL),
		bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send health check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health service returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}
