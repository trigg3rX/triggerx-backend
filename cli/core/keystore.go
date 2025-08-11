package core

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
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

// getPasswordSecurely prompts for password input without echoing to terminal
func getPasswordSecurely(prompt string) (string, error) {
	fmt.Print(prompt)

	// Get the file descriptor for stdin
	fd := int(syscall.Stdin)

	// Check if we're running in a terminal
	if term.IsTerminal(fd) {
		// Read password without echo
		bytePassword, err := term.ReadPassword(fd)
		fmt.Println() // Add newline after password input
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return strings.TrimSpace(string(bytePassword)), nil
	} else {
		// Fallback for non-terminal input (testing, etc.)
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return strings.TrimSpace(password), nil
	}
}

// confirmPassword prompts for password confirmation
func confirmPassword(password string) error {
	confirmPassword, err := getPasswordSecurely("Confirm password: ")
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}

// encryptBLS implements EIP-2335 encryption for BLS keystores
func encryptBLS(privateKeyBytes []byte, password string) (BLSKeystore, error) {
	// Generate random salt and IV
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return BLSKeystore{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return BLSKeystore{}, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Derive encryption key using PBKDF2
	dkLen := 32
	iterations := 262144
	key := pbkdf2.Key([]byte(password), salt, iterations, dkLen, sha256.New)

	// Split key into encryption key and MAC key
	encryptionKey := key[:16] // AES-128
	macKey := key[16:]

	// Encrypt private key using AES-128-CTR
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return BLSKeystore{}, fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(privateKeyBytes))
	stream.XORKeyStream(ciphertext, privateKeyBytes)

	// Calculate MAC
	macData := append(macKey, ciphertext...)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(macData)
	macSum := mac.Sum(nil)

	// Generate public key
	secretKey, err := bls.SecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		return BLSKeystore{}, fmt.Errorf("failed to create secret key: %w", err)
	}
	publicKey := secretKey.PublicKey().Marshal()

	// Create keystore structure
	keystore := BLSKeystore{
		Crypto: BLSCrypto{
			Cipher:     "aes-128-ctr",
			CipherText: hex.EncodeToString(ciphertext),
			CipherParams: map[string]string{
				"iv": hex.EncodeToString(iv),
			},
			KDF: "pbkdf2",
			KDFParams: map[string]interface{}{
				"dklen": dkLen,
				"c":     iterations,
				"prf":   "hmac-sha256",
				"salt":  hex.EncodeToString(salt),
			},
			MAC: hex.EncodeToString(macSum),
		},
		PubKey:  hex.EncodeToString(publicKey),
		Path:    "m/12381/3600/0/0",
		ID:      uuid.New().String(),
		Version: 4,
	}

	return keystore, nil
}

// decryptBLS implements EIP-2335 decryption for BLS keystores
func decryptBLS(keystore BLSKeystore, password string) ([]byte, error) {
	// Parse parameters
	salt, err := hex.DecodeString(keystore.Crypto.KDFParams["salt"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := hex.DecodeString(keystore.Crypto.CipherParams["iv"])
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := hex.DecodeString(keystore.Crypto.CipherText)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	expectedMAC, err := hex.DecodeString(keystore.Crypto.MAC)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MAC: %w", err)
	}

	// Derive key using same parameters
	dkLen := int(keystore.Crypto.KDFParams["dklen"].(float64))
	iterations := int(keystore.Crypto.KDFParams["c"].(float64))
	key := pbkdf2.Key([]byte(password), salt, iterations, dkLen, sha256.New)

	// Split key
	encryptionKey := key[:16]
	macKey := key[16:]

	// Verify MAC
	macData := append(macKey, ciphertext...)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(macData)
	calculatedMAC := mac.Sum(nil)

	if !hmac.Equal(expectedMAC, calculatedMAC) {
		return nil, fmt.Errorf("invalid password or corrupted keystore")
	}

	// Decrypt
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// GenerateBLSKeystore creates an encrypted BLS keystore in EIP-2335 format
func GenerateBLSKeystore(keystorePath string) (string, error) {
	// Generate random 32 bytes for the private key scalar
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Get password for encryption
	password, err := getPasswordSecurely("Enter password for BLS keystore: ")
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters long")
	}

	// Confirm password
	if err := confirmPassword(password); err != nil {
		return "", err
	}

	// Create encrypted keystore
	keystore, err := encryptBLS(privateKeyBytes, password)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt keystore: %w", err)
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

	// Return the private key hex for reference
	return hex.EncodeToString(privateKeyBytes), nil
}

// GenerateBLSKeystoreFromExistingKey creates an encrypted BLS keystore from an existing private key
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

	// Get password for encryption
	password, err := getPasswordSecurely("Enter password for BLS keystore: ")
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Confirm password
	if err := confirmPassword(password); err != nil {
		return err
	}

	// Create encrypted keystore
	keystore, err := encryptBLS(privateKeyBytes, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt keystore: %w", err)
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

// LoadBLSKeystore loads and decrypts a BLS keystore
// LoadBLSKeystore loads and decrypts a BLS keystore
func LoadBLSKeystore(keystorePath string, password string) (bls.SecretKey, error) {
	// Read keystore file
	keystoreData, err := os.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse keystore JSON
	var keystore BLSKeystore
	if err := json.Unmarshal(keystoreData, &keystore); err != nil {
		return nil, fmt.Errorf("failed to parse keystore: %w", err)
	}

	// Decrypt keystore
	secretKeyBytes, err := decryptBLS(keystore, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	// Create BLS secret key from decrypted bytes
	secretKey, err := bls.SecretKeyFromBytes(secretKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create BLS secret key: %w", err)
	}

	return secretKey, nil
}

// decryptECDSAKey decrypts an ECDSA keystore
func decryptECDSAKey(crypto Crypto, password string) ([]byte, error) {
	// Parse parameters
	salt, err := hex.DecodeString(crypto.KDFParams["salt"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := hex.DecodeString(crypto.CipherParams["iv"])
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := hex.DecodeString(crypto.CipherText)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	expectedMAC, err := hex.DecodeString(crypto.MAC)
	if err != nil {
		return nil, fmt.Errorf("failed to decode MAC: %w", err)
	}

	// Derive key using same parameters
	dkLen := int(crypto.KDFParams["dklen"].(float64))
	iterations := int(crypto.KDFParams["c"].(float64))
	key := pbkdf2.Key([]byte(password), salt, iterations, dkLen, sha256.New)

	// Split key
	encryptionKey := key[:16]
	macKey := key[16:]

	// Verify MAC
	macData := append(macKey, ciphertext...)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(macData)
	calculatedMAC := mac.Sum(nil)

	if !hmac.Equal(expectedMAC, calculatedMAC) {
		return nil, fmt.Errorf("invalid password or corrupted keystore")
	}

	// Decrypt
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// LoadECDSAKeystore loads and decrypts an ECDSA keystore
// LoadECDSAKeystore loads and decrypts an ECDSA keystore
func LoadECDSAKeystore(keystorePath string, password string) (*ecdsa.PrivateKey, error) {
	// Read keystore file
	keystoreData, err := os.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse keystore JSON
	var keystore Keystore
	if err := json.Unmarshal(keystoreData, &keystore); err != nil {
		return nil, fmt.Errorf("failed to parse keystore: %w", err)
	}

	// Decrypt keystore
	privateKeyBytes, err := decryptECDSAKey(keystore.Crypto, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	// Create ECDSA private key from decrypted bytes
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA private key: %w", err)
	}

	return privateKey, nil
}

// Simple PBKDF2-based encryption for ECDSA keystores
func encryptECDSAKey(privateKeyBytes []byte, password string) (Crypto, error) {
	// Generate random salt and IV
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return Crypto{}, fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)
	if err != nil {
		return Crypto{}, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Derive encryption key using PBKDF2
	dkLen := 32
	iterations := 262144
	key := pbkdf2.Key([]byte(password), salt, iterations, dkLen, sha256.New)

	// Split key into encryption key and MAC key
	encryptionKey := key[:16] // AES-128
	macKey := key[16:]

	// Encrypt private key using AES-128-CTR
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return Crypto{}, fmt.Errorf("failed to create cipher: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(privateKeyBytes))
	stream.XORKeyStream(ciphertext, privateKeyBytes)

	// Calculate MAC
	macData := append(macKey, ciphertext...)
	mac := hmac.New(sha256.New, macKey)
	mac.Write(macData)
	macSum := mac.Sum(nil)

	cryptoData := Crypto{
		Cipher:     "aes-128-ctr",
		CipherText: hex.EncodeToString(ciphertext),
		CipherParams: map[string]string{
			"iv": hex.EncodeToString(iv),
		},
		KDF: "pbkdf2",
		KDFParams: map[string]interface{}{
			"dklen": dkLen,
			"c":     iterations,
			"prf":   "hmac-sha256",
			"salt":  hex.EncodeToString(salt),
		},
		MAC: hex.EncodeToString(macSum),
	}

	return cryptoData, nil
}

// GenerateECDSAKeystore creates an encrypted ECDSA keystore file
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

	// Get password for encryption
	password, err := getPasswordSecurely("Enter password for ECDSA keystore: ")
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Confirm password
	if err := confirmPassword(password); err != nil {
		return err
	}

	// Encrypt the private key
	cryptoData, err := encryptECDSAKey(privateKeyBytes, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create keystore structure
	keystore := Keystore{
		Address: address.Hex()[2:], // Remove 0x prefix
		Crypto:  cryptoData,
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

	fmt.Printf("ECDSA keystore created successfully at: %s\n", keystorePath)
	fmt.Printf("Address: 0x%s\n", keystore.Address)

	return nil
}

// GenerateECDSAKeystoreFromExistingKey creates an encrypted ECDSA keystore from the private key in config
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

	// Get password for encryption
	password, err := getPasswordSecurely("Enter password for ECDSA keystore: ")
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Confirm password
	if err := confirmPassword(password); err != nil {
		return err
	}

	// Encrypt the private key
	cryptoData, err := encryptECDSAKey(privateKeyBytes, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create keystore structure
	keystore := map[string]interface{}{
		"address": address.Hex()[2:], // Remove 0x prefix
		"crypto":  cryptoData,
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

// GenerateBLSPublicKey generates a BLS public key from private key
func GenerateBLSPublicKey(privateKeyBytes []byte) []byte {
	secretKey, err := bls.SecretKeyFromBytes(privateKeyBytes)
	if err != nil {
		// Handle error (maybe return zero bytes or panic)
		return make([]byte, 48)
	}
	return secretKey.PublicKey().Marshal()
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
	// Note: Don't store the private key in plaintext in .env for encrypted keystores
	newContent += "# BLS_PRIVATE_KEY is encrypted in keystore - use keystore password to decrypt\n"

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
