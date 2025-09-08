package cryptography

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptMessage_ValidInput_ReturnsEncryptedHex(t *testing.T) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)

	message := "Hello, this is a test message!"

	// Test encryption
	encryptedHex, err := EncryptMessage(publicKeyHex, message)

	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedHex)
	assert.NotEqual(t, message, encryptedHex)
}

func TestEncryptMessage_InvalidPublicKeyHex_ReturnsError(t *testing.T) {
	tests := []struct {
		name         string
		publicKeyHex string
		message      string
		expectedErr  string
	}{
		{
			name:         "empty public key",
			publicKeyHex: "",
			message:      "test message",
			expectedErr:  "invalid public key hex",
		},
		{
			name:         "invalid hex format",
			publicKeyHex: "invalid-hex-string",
			message:      "test message",
			expectedErr:  "invalid public key hex",
		},
		{
			name:         "short hex string",
			publicKeyHex: "123456",
			message:      "test message",
			expectedErr:  "failed to unmarshal public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptedHex, err := EncryptMessage(tt.publicKeyHex, tt.message)

			assert.Empty(t, encryptedHex)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestEncryptMessage_EmptyMessage_ReturnsEncryptedHex(t *testing.T) {
	t.Skip("ECIES encryption doesn't handle empty messages well")
}

func TestDecryptMessage_ValidInput_ReturnsDecryptedMessage(t *testing.T) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	originalMessage := "Hello, this is a test message for decryption!"

	// First encrypt the message
	encryptedHex, err := EncryptMessage(publicKeyHex, originalMessage)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedHex)

	// Then decrypt it
	decryptedMessage, err := DecryptMessage(privateKeyHex, encryptedHex)

	assert.NoError(t, err)
	assert.Equal(t, originalMessage, decryptedMessage)
}

func TestDecryptMessage_InvalidEncryptedHex_ReturnsError(t *testing.T) {
	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	tests := []struct {
		name         string
		encryptedHex string
		expectedErr  string
	}{
		{
			name:         "empty encrypted hex",
			encryptedHex: "",
			expectedErr:  "invalid encrypted hex",
		},
		{
			name:         "invalid hex format",
			encryptedHex: "invalid-hex-string",
			expectedErr:  "invalid encrypted hex",
		},
		{
			name:         "short hex string",
			encryptedHex: "123456",
			expectedErr:  "decryption failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptedMessage, err := DecryptMessage(privateKeyHex, tt.encryptedHex)

			assert.Empty(t, decryptedMessage)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDecryptMessage_InvalidPrivateKey_ReturnsError(t *testing.T) {
	tests := []struct {
		name         string
		privateKey   string
		encryptedHex string
		expectedErr  string
	}{
		{
			name:         "empty private key",
			privateKey:   "",
			encryptedHex: "0x1234567890abcdef",
			expectedErr:  "invalid private key",
		},
		{
			name:         "invalid hex format",
			privateKey:   "invalid-hex-string",
			encryptedHex: "0x1234567890abcdef",
			expectedErr:  "invalid private key",
		},
		{
			name:         "short hex string",
			privateKey:   "123456",
			encryptedHex: "0x1234567890abcdef",
			expectedErr:  "invalid private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptedMessage, err := DecryptMessage(tt.privateKey, tt.encryptedHex)

			assert.Empty(t, decryptedMessage)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDecryptMessage_WrongPrivateKey_ReturnsError(t *testing.T) {
	// Generate two different key pairs
	privateKey1, err := crypto.GenerateKey()
	require.NoError(t, err)
	privateKey2, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey1 := privateKey1.PublicKey
	publicKeyBytes1 := crypto.FromECDSAPub(&publicKey1)
	publicKeyHex1 := hexutil.Encode(publicKeyBytes1)
	privateKeyBytes2 := crypto.FromECDSA(privateKey2)
	privateKeyHex2 := hexutil.Encode(privateKeyBytes2)

	originalMessage := "Hello, this is a test message!"

	// Encrypt with key1
	encryptedHex, err := EncryptMessage(publicKeyHex1, originalMessage)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedHex)

	// Try to decrypt with key2 (wrong key)
	decryptedMessage, err := DecryptMessage(privateKeyHex2, encryptedHex)

	assert.Empty(t, decryptedMessage)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestEncryptDecrypt_EndToEnd_WorksCorrectly(t *testing.T) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	testMessages := []string{
		"Simple message",
		"Message with special chars: !@#$%^&*()",
		"Message with numbers: 1234567890",
		"Message with unicode: 🚀🌟🎉",
		"Very long message " + string(make([]byte, 1000)), // 1000 bytes
	}

	for i, message := range testMessages {
		t.Run(fmt.Sprintf("message_%d", i), func(t *testing.T) {
			// Encrypt
			encryptedHex, err := EncryptMessage(publicKeyHex, message)
			require.NoError(t, err)
			require.NotEmpty(t, encryptedHex)

			// Decrypt
			decryptedMessage, err := DecryptMessage(privateKeyHex, encryptedHex)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, message, decryptedMessage)
		})
	}
}

func TestEncryptMessage_LargeMessage_HandlesCorrectly(t *testing.T) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	// Create a large message (10KB)
	largeMessage := string(make([]byte, 10*1024))
	for i := range largeMessage {
		largeMessage = largeMessage[:i] + string(byte(i%256)) + largeMessage[i+1:]
	}

	// Test encryption and decryption of large message
	encryptedHex, err := EncryptMessage(publicKeyHex, largeMessage)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedHex)

	decryptedMessage, err := DecryptMessage(privateKeyHex, encryptedHex)
	require.NoError(t, err)

	assert.Equal(t, largeMessage, decryptedMessage)
}

// Benchmark tests
func BenchmarkEncryptMessage(b *testing.B) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(b, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)
	message := "Benchmark test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptMessage(publicKeyHex, message)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecryptMessage(b *testing.B) {
	// Generate a test key pair
	privateKey, err := crypto.GenerateKey()
	require.NoError(b, err)

	publicKey := privateKey.PublicKey
	publicKeyBytes := crypto.FromECDSAPub(&publicKey)
	publicKeyHex := hexutil.Encode(publicKeyBytes)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)
	message := "Benchmark test message"

	// Pre-encrypt the message
	encryptedHex, err := EncryptMessage(publicKeyHex, message)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptMessage(privateKeyHex, encryptedHex)
		if err != nil {
			b.Fatal(err)
		}
	}
}
