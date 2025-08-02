package proof

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TLSProofConfig holds configuration for TLS proof generation
type TLSProofConfig struct {
	TargetHost string        // Host to establish TLS connection with
	TargetPort string        // Port to connect to (default: "443")
	Timeout    time.Duration // Connection timeout (default: 10s)
	VerifyPeer bool          // Whether to verify peer certificates (default: true)
	ServerName string        // Server name for SNI (optional, defaults to TargetHost)
}

// DefaultTLSProofConfig returns a default configuration for TLS proof generation
func DefaultTLSProofConfig(host string) *TLSProofConfig {
	return &TLSProofConfig{
		TargetHost: host,
		TargetPort: "443",
		Timeout:    5 * time.Second,
		VerifyPeer: true,
		ServerName: host,
	}
}

// GenerateProofWithTLSConnection generates a proof using a real TLS connection
func GenerateProofWithTLSConnection(ipfsData types.IPFSData, config *TLSProofConfig) (types.ProofData, error) {
	connState, err := EstablishTLSConnection(config)
	if err != nil {
		return types.ProofData{}, fmt.Errorf("failed to establish TLS connection: %w", err)
	}

	return GenerateProof(ipfsData, connState)
}

// EstablishTLSConnection creates a real TLS connection and returns the connection state
func EstablishTLSConnection(config *TLSProofConfig) (*tls.ConnectionState, error) {
	if config == nil {
		return nil, errors.New("TLS proof config cannot be nil")
	}

	if config.TargetHost == "" {
		return nil, errors.New("target host cannot be empty")
	}

	if config.TargetPort == "" {
		config.TargetPort = "443"
	}

	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	if config.ServerName == "" {
		config.ServerName = config.TargetHost
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		ServerName:         config.ServerName,
		InsecureSkipVerify: !config.VerifyPeer,
	}

	// Establish connection with timeout
	dialer := &net.Dialer{
		Timeout: config.Timeout,
	}

	address := net.JoinHostPort(config.TargetHost, config.TargetPort)
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish TLS connection to %s: %w", address, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Warning: failed to close connection: %v\n", err)
		}
	}()

	// Get connection state
	connState := conn.ConnectionState()

	// Verify we have certificates
	if len(connState.PeerCertificates) == 0 {
		return nil, errors.New("no peer certificates found in TLS connection")
	}

	return &connState, nil
}

// GenerateProof takes the action execution data and generates a proof by:
// 1. Creating a hash of the action data (tx hash, gas used, status etc)
// 2. Adding a timestamp when the proof was generated
// 3. Including TLS certificate information for verification
// This provides cryptographic proof that the action was executed
func GenerateProof(ipfsData types.IPFSData, connState *tls.ConnectionState) (types.ProofData, error) {
	// Validate inputs
	if connState == nil {
		return types.ProofData{}, errors.New("TLS connection state cannot be nil")
	}

	if len(connState.PeerCertificates) == 0 {
		return types.ProofData{}, errors.New("no TLS certificates found in connection state")
	}

	// Stringify the ipfs data, after converting all the strings to lowercase
	dataStr, err := StringifyIPFSData(ipfsData)
	if err != nil {
		return types.ProofData{}, fmt.Errorf("failed to stringify IPFS data: %w", err)
	}

	// Validate and process the first certificate
	cert := connState.PeerCertificates[0]
	if cert == nil {
		return types.ProofData{}, errors.New("first peer certificate is nil")
	}

	// Verify certificate validity
	if err := validateCertificate(cert); err != nil {
		return types.ProofData{}, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Hash the certificate
	certHash := sha256.Sum256(cert.Raw)
	certHashStr := hex.EncodeToString(certHash[:])

	// Generate proof hash from the data
	proofHash := sha256.Sum256([]byte(dataStr))
	proofHashStr := hex.EncodeToString(proofHash[:])

	// Create enhanced proof with additional TLS information
	proofData := types.ProofData{
		TaskID:               ipfsData.TaskData.TaskID[0],
		ProofOfTask:          proofHashStr,
		CertificateHash:      certHashStr,
		CertificateTimestamp: time.Now().UTC(),
	}

	return proofData, nil
}

// validateCertificate performs basic validation on the certificate
func validateCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("certificate is nil")
	}

	// Check if certificate is expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from: %v)", cert.NotBefore)
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired on: %v)", cert.NotAfter)
	}

	// Check if certificate has required fields
	if len(cert.Raw) == 0 {
		return errors.New("certificate raw data is empty")
	}

	return nil
}

func StringifyIPFSData(ipfsData types.IPFSData) (string, error) {
	ipfsDataMap := make(map[string]interface{})
	ipfsDataMap["task_data"] = ipfsData.TaskData
	ipfsDataMap["action_data"] = ipfsData.ActionData
	ipfsDataMap["proof_data"] = ipfsData.ProofData
	ipfsDataMap["performer_signature"] = ipfsData.PerformerSignature

	convertToLower(ipfsDataMap)

	ipfsDataBytes, err := json.Marshal(ipfsDataMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IPFS data: %w", err)
	}
	return string(ipfsDataBytes), nil
}

func convertToLower(data map[string]interface{}) {
	for k, v := range data {
		if s, ok := v.(string); ok {
			data[k] = strings.ToLower(s)
		} else if m, ok := v.(map[string]interface{}); ok {
			convertToLower(m)
		}
	}
}
