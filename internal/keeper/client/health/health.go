package health

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/cryptography"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
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
	httpClient *httppkg.HTTPClient
	logger     observability.Logger
	tracer     observability.Tracer
	config     Config
}

// Config holds the configuration for the Health client
type Config struct {
	HealthServiceURL string
	PrivateKey       string
	KeeperAddress    string
	PeerID           string
	Version          string
	Network          string
	RequestTimeout   time.Duration
}

// NewClient creates a new Health service client
func NewClient(logger observability.Logger, tracer observability.Tracer, cfg Config) (*Client, error) {
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10 * time.Second
	}

	if cfg.Version == "" {
		cfg.Version = config.GetVersion()
	}

	// Configure retry policy: 3 retries for health check-in
	retryConfig := httppkg.DefaultHTTPRetryConfig()
	retryConfig.RetryConfig.MaxRetries = 3

	httpClient, err := httppkg.NewHTTPClient(retryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &Client{
		httpClient: httpClient,
		logger:     logger,
		tracer:     tracer,
		config:     cfg,
	}, nil
}

// CheckIn performs a health check-in with the health service
func (c *Client) CheckIn(ctx context.Context) (types.KeeperHealthCheckInResponse, error) {
	// Start a span for the health check-in operation
	var span observability.Span
	if c.tracer != nil {
		ctx, span = c.tracer.Start(ctx, "keeper.health_check_in",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				attribute.String("keeper.address", c.config.KeeperAddress),
				attribute.String("keeper.version", c.config.Version),
				attribute.String("keeper.peer_id", c.config.PeerID),
			),
		)
		defer span.End()
	}

	// Get consensus address from private key
	privateKey, err := ethcrypto.HexToECDSA(c.config.PrivateKey)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
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
	signature, err := cryptography.SignMessage(string(msg), c.config.PrivateKey)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
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
		IsImua:           config.IsImua(),
		Network:          c.config.Network,
	}

	// c.logger.Info(ctx, "Payload", observability.Any("payload", payload))

	// Send health check request
	response, err := c.sendHealthCheck(ctx, payload)
	if err != nil {
		if metrics.FailedHealthCheckinsTotal != nil {
			metrics.FailedHealthCheckinsTotal.Inc(ctx)
		}
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("health check failed: %w", err)
	}

	// Track success or failure based on response status
	if response.Status {
		if span != nil {
			span.SetAttributes(
				attribute.Bool("health_check.success", true),
			)
			span.SetStatus(codes.Ok, "")
		}
	} else {
		if metrics.FailedHealthCheckinsTotal != nil {
			metrics.FailedHealthCheckinsTotal.Inc(ctx)
		}
		if span != nil {
			span.SetAttributes(
				attribute.Bool("health_check.success", false),
			)
			span.SetStatus(codes.Error, "health check returned failure status")
		}
	}

	// c.logger.Debug("Successfully completed health check-in",
	// 	"status", response.Status,
	// 	"keeperAddress", c.config.KeeperAddress,
	// 	"timestamp", payload.Timestamp)

	return response, nil
}

// sendHealthCheck sends the health check request to the health service
func (c *Client) sendHealthCheck(ctx context.Context, payload types.KeeperHealthCheckIn) (types.KeeperHealthCheckInResponse, error) {
	// Start a span for the HTTP request
	var span observability.Span
	if c.tracer != nil {
		ctx, span = c.tracer.Start(ctx, "keeper.health_check_in.http_request",
			observability.WithSpanKind(trace.SpanKindClient),
			observability.WithAttributes(
				semconv.HTTPMethodKey.String("POST"),
				semconv.HTTPURLKey.String(fmt.Sprintf("%s/health", c.config.HealthServiceURL)),
			),
		)
		defer span.End()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to marshal health check payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/health", c.config.HealthServiceURL),
		bytes.NewBuffer(payloadBytes))
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Generate descriptive trace ID: post-health-{keeperaddress}-{randomness}
	traceID := generateHealthCheckTraceID(c.config.KeeperAddress)
	req.Header.Set("X-Trace-ID", traceID)

	// Add trace ID to span attributes for visibility
	if span != nil {
		span.SetAttributes(attribute.String("trace.id", traceID))
	}

	// Inject trace context into HTTP headers (OpenTelemetry will use the X-Trace-ID if present)
	if c.tracer != nil {
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	}

	resp, err := c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to send health check request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn(ctx, "failed to close response body", observability.Error(err))
		}
	}()

	// Set HTTP response attributes on span
	if span != nil {
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(resp.StatusCode),
		)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			if errResp.Code == "KEEPER_NOT_VERIFIED" {
				if span != nil {
					span.SetStatus(codes.Error, "keeper not verified")
					span.SetAttributes(attribute.String("error.code", errResp.Code))
				}
				return types.KeeperHealthCheckInResponse{
					Status: false,
					Data:   errResp.Error,
				}, ErrKeeperNotVerified
			}
		}
		if span != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   errResp.Error,
		}, fmt.Errorf("health service returned non-OK status: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var response types.KeeperHealthCheckInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return types.KeeperHealthCheckInResponse{
			Status: false,
			Data:   err.Error(),
		}, fmt.Errorf("failed to unmarshal health check response: %w", err)
	}

	// Only decrypt if the response was successful
	if response.Status {
		decryptedString, err := cryptography.DecryptMessage(c.config.PrivateKey, response.Data)
		if err != nil {
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return types.KeeperHealthCheckInResponse{
				Status: false,
				Data:   err.Error(),
			}, fmt.Errorf("failed to decrypt health check response: %w", err)
		}

		parts := strings.Split(decryptedString, ":")
		if len(parts) != 6 {
			err := fmt.Errorf("invalid response format: expected 6 parts")
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return types.KeeperHealthCheckInResponse{
				Status: false,
				Data:   "invalid response format",
			}, err
		}

		config.SetEtherscanAPIKey(parts[0])
		config.SetAlchemyAPIKey(parts[1])
		config.SetIpfsHost(parts[2])
		config.SetPinataJWT(parts[3])
		config.SetManagerSigningAddress(parts[4])
		config.SetTaskExecutionAddress(parts[5])
		// config.SetTaskExecutionAddress("0x3509F38e10eB3cDcE7695743cB7e81446F4d8A33")

		if span != nil {
			span.SetStatus(codes.Ok, "")
		}

		return types.KeeperHealthCheckInResponse{
			Status: true,
			Data:   "Health check-in successful",
		}, nil
	}

	// If response was not successful, return the error as is
	if span != nil {
		span.SetStatus(codes.Error, "health check-in failed")
	}
	return response, nil
}

// generateHealthCheckTraceID generates a descriptive trace ID for health check-ins
// Format: post-health-{keeperaddress}-{randomness}
func generateHealthCheckTraceID(keeperAddress string) string {
	// Normalize keeper address: remove 0x prefix and convert to lowercase
	normalizedAddr := strings.ToLower(keeperAddress)
	normalizedAddr = strings.TrimPrefix(normalizedAddr, "0x")

	// Generate 8 bytes of randomness for uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based randomness if crypto/rand fails
		randomBytes = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}
	randomness := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("post-health-%s-%s", normalizedAddr, randomness)
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.httpClient.Close()
}
