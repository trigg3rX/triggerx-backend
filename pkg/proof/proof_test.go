package proof

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func TestGenerateProofWithTLSConnection(t *testing.T) {
	// Create sample IPFS data for testing
	ipfsData := types.IPFSData{
		TaskData: &types.SendTaskDataToKeeper{
			TaskID: []int64{123},
		},
		ActionData: &types.PerformerActionData{
			TaskID:             123,
			ActionTxHash:       "0x1234567890abcdef",
			GasUsed:            "21000",
			Status:             true,
			ExecutionTimestamp: time.Now(),
		},
	}

	// Test with Google (reliable TLS endpoint)
	config := DefaultTLSProofConfig("www.google.com")

	proofData, err := GenerateProofWithTLSConnection(ipfsData, config)
	if err != nil {
		t.Fatalf("Failed to generate TLS proof: %v", err)
	}

	// Validate proof data
	if proofData.TaskID != ipfsData.TaskData.TaskID[0] {
		t.Errorf("Expected TaskID %d, got %d", ipfsData.TaskData.TaskID, proofData.TaskID)
	}

	if proofData.ProofOfTask == "" {
		t.Error("ProofOfTask should not be empty")
	}

	if proofData.CertificateHash == "" {
		t.Error("CertificateHash should not be empty")
	}

	if proofData.CertificateTimestamp.IsZero() {
		t.Error("CertificateTimestamp should not be zero")
	}

	t.Logf("Successfully generated TLS proof: %+v", proofData)
}

func TestGenerateProofWithMockCert(t *testing.T) {
	// Create sample IPFS data for testing
	ipfsData := types.IPFSData{
		TaskData: &types.SendTaskDataToKeeper{
			TaskID: []int64{456},
		},
		ActionData: &types.PerformerActionData{
			TaskID:             456,
			ActionTxHash:       "0xabcdef1234567890",
			GasUsed:            "42000",
			Status:             true,
			ExecutionTimestamp: time.Now(),
		},
	}

	// Create a mock certificate with proper structure
	mockCert := &x509.Certificate{
		Raw:       []byte("mock_certificate_data_for_testing_only"),
		NotBefore: time.Now().Add(-24 * time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
	}

	connState := &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{mockCert},
	}

	proofData, err := GenerateProof(ipfsData, connState)
	if err != nil {
		t.Fatalf("Failed to generate mock proof: %v", err)
	}

	// Validate proof data
	if proofData.TaskID != ipfsData.TaskData.TaskID[0] {
		t.Errorf("Expected TaskID %d, got %d", ipfsData.TaskData.TaskID, proofData.TaskID)
	}

	if proofData.ProofOfTask == "" {
		t.Error("ProofOfTask should not be empty")
	}

	if proofData.CertificateHash == "" {
		t.Error("CertificateHash should not be empty")
	}

	t.Logf("Successfully generated mock proof: %+v", proofData)
}

func TestTLSProofConfig(t *testing.T) {
	// Test default config
	config := DefaultTLSProofConfig("example.com")

	if config.TargetHost != "example.com" {
		t.Errorf("Expected TargetHost 'example.com', got '%s'", config.TargetHost)
	}

	if config.TargetPort != "443" {
		t.Errorf("Expected TargetPort '443', got '%s'", config.TargetPort)
	}

	if config.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout 10s, got %v", config.Timeout)
	}

	if !config.VerifyPeer {
		t.Error("Expected VerifyPeer to be true")
	}

	if config.ServerName != "example.com" {
		t.Errorf("Expected ServerName 'example.com', got '%s'", config.ServerName)
	}
}

func TestEstablishTLSConnection(t *testing.T) {
	// Test with a reliable host
	config := DefaultTLSProofConfig("www.google.com")

	connState, err := EstablishTLSConnection(config)
	if err != nil {
		t.Fatalf("Failed to establish TLS connection: %v", err)
	}

	if connState == nil {
		t.Fatal("Connection state should not be nil")
	}

	if len(connState.PeerCertificates) == 0 {
		t.Fatal("Should have peer certificates")
	}

	t.Logf("Successfully established TLS connection with %d certificates", len(connState.PeerCertificates))
}

func TestEstablishTLSConnectionWithInvalidConfig(t *testing.T) {
	// Test with nil config
	_, err := EstablishTLSConnection(nil)
	if err == nil {
		t.Error("Expected error with nil config")
	}

	// Test with empty host
	config := &TLSProofConfig{}
	_, err = EstablishTLSConnection(config)
	if err == nil {
		t.Error("Expected error with empty host")
	}
}
