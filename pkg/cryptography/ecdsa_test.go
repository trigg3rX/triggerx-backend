package cryptography

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data - using a known private key for consistent testing
const (
	testPrivateKey = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testMessage    = "Hello, Ethereum!"
)

func TestSignMessage_ValidInput_ReturnsSignature(t *testing.T) {
	// Test with valid private key and message
	signature, err := SignMessage(testMessage, testPrivateKey)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 132) // 0x + 130 hex chars = 132 total
}

func TestSignMessage_InvalidPrivateKey_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		privateKey  string
		expectedErr string
	}{
		{"empty private key", "", "invalid private key"},
		{"invalid hex", "invalid-hex", "invalid private key"},
		{"too short", "123456", "invalid private key"},
		{"too long", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "invalid private key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SignMessage(testMessage, tt.privateKey)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestSignMessage_EmptyMessage_ReturnsSignature(t *testing.T) {
	signature, err := SignMessage("", testPrivateKey)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestSignJSONMessage_ValidJSON_ReturnsSignature(t *testing.T) {
	testData := map[string]interface{}{
		"name":    "Test User",
		"age":     25,
		"active":  true,
		"address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
	}

	signature, err := SignJSONMessage(testData, testPrivateKey)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestSignJSONMessage_ComplexJSON_ReturnsSignature(t *testing.T) {
	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "JOHN@EXAMPLE.COM",
		},
		"settings": map[string]interface{}{
			"theme":         "DARK",
			"notifications": true,
		},
	}

	signature, err := SignJSONMessage(testData, testPrivateKey)

	require.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestSignJSONMessage_InvalidJSON_ReturnsError(t *testing.T) {
	// Create a channel which cannot be marshaled to JSON
	invalidData := make(chan int)

	_, err := SignJSONMessage(invalidData, testPrivateKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal input data")
}

func TestVerifySignature_ValidSignature_ReturnsTrue(t *testing.T) {
	// Generate a signature first
	signature, err := SignMessage(testMessage, testPrivateKey)
	require.NoError(t, err)

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify the signature
	isValid, err := VerifySignature(testMessage, signature, address.Hex())

	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestVerifySignature_InvalidSignature_ReturnsFalse(t *testing.T) {
	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Test with signature from a different message
	differentMessage := "Different message"
	wrongSignature, err := SignMessage(differentMessage, testPrivateKey)
	require.NoError(t, err)

	// Verify the signature against the original message (should be false)
	isValid, err := VerifySignature(testMessage, wrongSignature, address.Hex())

	require.NoError(t, err)
	assert.False(t, isValid)
}

func TestVerifySignature_InvalidSignatureFormat_ReturnsError(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		expectedError string
	}{
		{"empty signature", "", "invalid signature"},
		{"invalid hex", "invalid-hex", "invalid signature"},
		{"too short", "0x123456", "invalid signature length"},
		{"too long", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "invalid signature length"},
	}

	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifySignature(testMessage, tt.signature, address.Hex())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestVerifySignatureFromJSON_ValidSignature_ReturnsTrue(t *testing.T) {
	// Test data with mixed case that should be converted to lower
	testData := map[string]interface{}{
		"name":    "Test User",
		"age":     25,
		"active":  true,
		"address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
		"email":   "TEST@EXAMPLE.COM",
	}

	// Sign the JSON data
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify the signature
	isValid, err := VerifySignatureFromJSON(testData, signature, address.Hex())

	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestVerifySignatureFromJSON_ComplexJSON_ReturnsTrue(t *testing.T) {
	// Test with nested JSON structure
	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "John Doe",
			"email": "JOHN@EXAMPLE.COM",
		},
		"settings": map[string]interface{}{
			"theme":         "DARK",
			"notifications": true,
		},
		"metadata": map[string]interface{}{
			"version": "1.0.0",
			"status":  "ACTIVE",
		},
	}

	// Sign the JSON data
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify the signature
	isValid, err := VerifySignatureFromJSON(testData, signature, address.Hex())

	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestVerifySignatureFromJSON_InvalidSignature_ReturnsFalse(t *testing.T) {
	testData := map[string]interface{}{
		"name":    "Test User",
		"age":     25,
		"active":  true,
		"address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
	}

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Test with signature from a different message
	differentData := map[string]interface{}{
		"name": "Different User",
		"age":  30,
	}
	wrongSignature, err := SignJSONMessage(differentData, testPrivateKey)
	require.NoError(t, err)

	// Verify the signature against the original data (should be false)
	isValid, err := VerifySignatureFromJSON(testData, wrongSignature, address.Hex())

	require.NoError(t, err)
	assert.False(t, isValid)
}

func TestVerifySignatureFromJSON_WrongSignerAddress_ReturnsFalse(t *testing.T) {
	testData := map[string]interface{}{
		"name":    "Test User",
		"age":     25,
		"active":  true,
		"address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
	}

	// Sign the JSON data
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	// Use a different address (wrong signer)
	wrongAddress := "0x1234567890123456789012345678901234567890"

	// Verify the signature with wrong address (should be false)
	isValid, err := VerifySignatureFromJSON(testData, signature, wrongAddress)

	require.NoError(t, err)
	assert.False(t, isValid)
}

func TestVerifySignatureFromJSON_InvalidSignatureFormat_ReturnsError(t *testing.T) {
	tests := []struct {
		name          string
		signature     string
		expectedError string
	}{
		{"empty signature", "", "invalid signature"},
		{"invalid hex", "invalid-hex", "invalid signature"},
		{"too short", "0x123456", "invalid signature length"},
		{"too long", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "invalid signature length"},
	}

	testData := map[string]interface{}{
		"name": "Test User",
		"age":  25,
	}

	// Get a valid address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := VerifySignatureFromJSON(testData, tt.signature, address.Hex())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestVerifySignatureFromJSON_InvalidJSON_ReturnsError(t *testing.T) {
	// Create a channel which cannot be marshaled to JSON
	invalidData := make(chan int)

	// Get a valid signature and address for testing
	testData := map[string]interface{}{
		"name": "Test User",
	}
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Test with invalid JSON data
	_, err = VerifySignatureFromJSON(invalidData, signature, address.Hex())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal input data")
}

func TestVerifySignatureFromJSON_EmptyJSON_ReturnsTrue(t *testing.T) {
	// Test with empty JSON object
	testData := map[string]interface{}{}

	// Sign the JSON data
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify the signature
	isValid, err := VerifySignatureFromJSON(testData, signature, address.Hex())

	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestVerifySignatureFromJSON_NonMapJSON_ReturnsError(t *testing.T) {
	// Test with non-map JSON data (array, string, number, etc.) which should fail
	testCases := []interface{}{
		"simple string",
		42,
		true,
		[]interface{}{"item1", "item2"},
		[]interface{}{1, 2, 3},
	}

	for _, testData := range testCases {
		t.Run(fmt.Sprintf("%T", testData), func(t *testing.T) {
			// Get a valid signature and address for testing
			validData := map[string]interface{}{
				"name": "Test User",
			}
			signature, err := SignJSONMessage(validData, testPrivateKey)
			require.NoError(t, err)

			privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
			require.NoError(t, err)

			publicKey := privateKeyECDSA.Public()
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			require.True(t, ok)

			address := crypto.PubkeyToAddress(*publicKeyECDSA)

			// Test with non-map JSON data (should fail at unmarshal step)
			_, err = VerifySignatureFromJSON(testData, signature, address.Hex())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to unmarshal to map")
		})
	}
}

func TestVerifySignatureFromJSON_CaseInsensitiveStringConversion(t *testing.T) {
	// Test that string values are converted to lowercase during verification
	testData := map[string]interface{}{
		"name":    "Test User",
		"email":   "TEST@EXAMPLE.COM",
		"status":  "ACTIVE",
		"country": "UNITED STATES",
	}

	// Sign the JSON data (this will convert strings to lowercase)
	signature, err := SignJSONMessage(testData, testPrivateKey)
	require.NoError(t, err)

	// Get the corresponding public key/address
	privateKeyECDSA, err := crypto.HexToECDSA(testPrivateKey)
	require.NoError(t, err)

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Verify with the same data (should work because both signing and verification convert to lowercase)
	isValid, err := VerifySignatureFromJSON(testData, signature, address.Hex())

	require.NoError(t, err)
	assert.True(t, isValid)
}
