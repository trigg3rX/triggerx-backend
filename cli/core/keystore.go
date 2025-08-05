package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/trigg3rX/triggerx-backend/cli/core/config"
)

// BLS EIP-2335 keystore structure
type BLSCrypto struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams map[string]string      `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type BLSKeystore struct {
	Crypto  BLSCrypto `json:"crypto"`
	PubKey  string    `json:"pubkey"`
	Path    string    `json:"path"`
	ID      string    `json:"uuid"`
	Version int       `json:"version"`
}

// ECDSA keystore structure
type Crypto struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams map[string]string      `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type Keystore struct {
	Address string `json:"address,omitempty"`
	Crypto  Crypto `json:"crypto"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

// GenerateBLSKeystore creates a BLS keystore in EIP-2335 format
func GenerateBLSKeystore(keystorePath string) (string, error) {
	// Generate random 32 bytes for the private key scalar
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string for the private key
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Generate corresponding public key
	publicKeyBytes := GenerateBLSPublicKey(privateKeyBytes)

	// Create keystore
	keystore := map[string]interface{}{
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": privateKeyHex, // Store private key as plaintext hex
			"cipherparams": map[string]string{
				"iv": "00000000000000000000000000000000", // All zeros
			},
			"kdf": "pbkdf2",
			"kdfparams": map[string]interface{}{
				"dklen": 32,
				"c":     1, // Low iteration count
				"prf":   "hmac-sha256",
				"salt":  "0000000000000000000000000000000000000000000000000000000000000000", // All zeros
			},
			"mac": "0000000000000000000000000000000000000000000000000000000000000000", // All zeros
		},
		"pubkey":  hex.EncodeToString(publicKeyBytes),
		"path":    "m/12381/3600/0/0",
		"uuid":    uuid.New().String(),
		"version": 4,
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to write keystore file: %w", err)
	}

	return privateKeyHex, nil
}

// GenerateBLSKeystoreFromExistingKey creates a BLS keystore from an existing private key
func GenerateBLSKeystoreFromExistingKey(keystorePath, privateKeyHex string) error {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Decode private key bytes
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to decode private key hex: %w", err)
	}

	// Generate corresponding public key
	publicKeyBytes := GenerateBLSPublicKey(privateKeyBytes)

	// Create keystore structure
	keystore := map[string]interface{}{
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": privateKeyHex,
			"cipherparams": map[string]string{
				"iv": "00000000000000000000000000000000",
			},
			"kdf": "pbkdf2",
			"kdfparams": map[string]interface{}{
				"dklen": 32,
				"c":     1,
				"prf":   "hmac-sha256",
				"salt":  "0000000000000000000000000000000000000000000000000000000000000000",
			},
			"mac": "0000000000000000000000000000000000000000000000000000000000000000",
		},
		"pubkey":  hex.EncodeToString(publicKeyBytes),
		"path":    "m/12381/3600/0/0",
		"uuid":    uuid.New().String(),
		"version": 4,
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// GenerateBLSPublicKey generates a BLS public key from private key
func GenerateBLSPublicKey(privateKeyBytes []byte) []byte {
	secretKey, err := bls.SecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		// Handle error (maybe return zero bytes or panic)
		return make([]byte, 48)
	}
	return secretKey.PublicKey().Marshal()
}

// GenerateECDSAKeystore creates an ECDSA keystore file
func GenerateECDSAKeystore(keystorePath string) error {
	// Generate new ECDSA private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Get address from public key
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	// Simple keystore structure
	keystore := Keystore{
		Address: address.Hex()[2:], // Remove 0x prefix
		Crypto: Crypto{
			Cipher:     "aes-128-ctr",
			CipherText: hex.EncodeToString(privateKeyBytes), // Storing unencrypted for simplicity
			CipherParams: map[string]string{
				"iv": hex.EncodeToString(iv),
			},
			KDF: "scrypt",
			KDFParams: map[string]interface{}{
				"dklen": 32,
				"n":     262144,
				"p":     1,
				"r":     8,
				"salt":  hex.EncodeToString(salt),
			},
			MAC: "0000000000000000000000000000000000000000000000000000000000000000", // Dummy MAC
		},
		ID:      uuid.New().String(),
		Version: 3,
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// GenerateECDSAKeystoreFromExistingKey creates an ECDSA keystore from the private key in config
func GenerateECDSAKeystoreFromExistingKey(keystorePath string) error {
	// Get the ECDSA private key from config
	privateKey := config.GetEcdsaPrivateKey()
	if privateKey == nil {
		return fmt.Errorf("ECDSA private key not available from config")
	}

	// Get address from public key
	address := config.GetOperatorAddress()

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	// Create an unencrypted keystore
	keystore := map[string]interface{}{
		"address": address.Hex()[2:], // Remove 0x prefix
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": hex.EncodeToString(privateKeyBytes),
			"cipherparams": map[string]string{
				"iv": "00000000000000000000000000000000", // All zeros IV for no encryption
			},
			"kdf": "pbkdf2",
			"kdfparams": map[string]interface{}{
				"dklen": 32,
				"c":     1, // Low iteration count for no password
				"prf":   "hmac-sha256",
				"salt":  "0000000000000000000000000000000000000000000000000000000000000000", // All zeros salt
			},
			"mac": "0000000000000000000000000000000000000000000000000000000000000000", // All zeros MAC to indicate no encryption verification
		},
		"id":      uuid.New().String(),
		"version": 3,
	}

	// Write keystore to file
	keystoreJSON, err := json.MarshalIndent(keystore, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keystore: %w", err)
	}

	err = os.WriteFile(keystorePath, keystoreJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// UpdateEnvFile updates the .env file with keystore paths
func UpdateEnvFile(envPath, ecdsaKeyPath, blsKeyPath, blsPrivateKeyHex string) error {
	// Read existing .env file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Add keystore paths
	newContent := string(content)
	newContent += "\n# Keystore paths (generated by generate-keys command)\n"
	newContent += fmt.Sprintf("ECDSA_PRIVATE_KEY_STORE_PATH=%s\n", ecdsaKeyPath)
	newContent += fmt.Sprintf("BLS_PRIVATE_KEY_STORE_PATH=%s\n", blsKeyPath)
	newContent += fmt.Sprintf("BLS_PRIVATE_KEY=%s\n", blsPrivateKeyHex)

	// Write back to file
	err = os.WriteFile(envPath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// UpdateEnvFileWithPathsAndAddress updates the .env file with keystore paths and operator address
func UpdateEnvFileWithPathsAndAddress(ecdsaKeyPath, blsKeyPath, operatorAddress string) error {
	envPath := ".env"

	// Read existing .env file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	lines := splitLines(string(content))
	var newLines []string
	var addedECDSAPath, addedBLSPath, addedOperatorAddr bool

	// Process existing lines and update if they exist
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "ECDSA_PRIVATE_KEY_STORE_PATH=") {
			newLines = append(newLines, fmt.Sprintf("ECDSA_PRIVATE_KEY_STORE_PATH=%s", ecdsaKeyPath))
			addedECDSAPath = true
		} else if strings.HasPrefix(trimmedLine, "BLS_PRIVATE_KEY_STORE_PATH=") {
			newLines = append(newLines, fmt.Sprintf("BLS_PRIVATE_KEY_STORE_PATH=%s", blsKeyPath))
			addedBLSPath = true
		} else if strings.HasPrefix(trimmedLine, "OPERATOR_ADDRESS=") {
			newLines = append(newLines, fmt.Sprintf("OPERATOR_ADDRESS=%s", operatorAddress))
			addedOperatorAddr = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Add missing variables at the end
	if !addedECDSAPath || !addedBLSPath || !addedOperatorAddr {
		newLines = append(newLines, "\n# Keystore paths and operator address (generated by complete-registration command)")
		if !addedECDSAPath {
			newLines = append(newLines, fmt.Sprintf("ECDSA_PRIVATE_KEY_STORE_PATH=%s", ecdsaKeyPath))
		}
		if !addedBLSPath {
			newLines = append(newLines, fmt.Sprintf("BLS_PRIVATE_KEY_STORE_PATH=%s", blsKeyPath))
		}
		if !addedOperatorAddr {
			newLines = append(newLines, fmt.Sprintf("OPERATOR_ADDRESS=%s", operatorAddress))
		}
	}

	// Write back to file
	envContent := strings.Join(newLines, "\n")
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// Helper functions
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
