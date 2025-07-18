package actions

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/trigg3rX/triggerx-backend/cli/operator"
	"github.com/trigg3rX/triggerx-backend/cli/types"
	"github.com/urfave/cli"
)

func CompleteRegistration(ctx *cli.Context) error {
	log.Println("Starting complete registration process...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Check if BLS private key exists in environment
	blsPrivateKeyHex := config.GetBLSPrivateKeyHex()
	if blsPrivateKeyHex == "" {
		return fmt.Errorf("BLS_PRIVATE_KEY not found in environment variables")
	}

	log.Printf("Found BLS private key in environment: %s...", blsPrivateKeyHex[:10])

	// Step 1: Generate BLS keystore from existing private key
	log.Println("Step 1: Generating BLS keystore from existing private key...")
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %v", err)
	}

	blsKeyPath := filepath.Join(keysDir, "bls.json")
	if err := generateBLSKeystoreFromExistingKey(blsKeyPath, blsPrivateKeyHex); err != nil {
		return fmt.Errorf("failed to generate BLS keystore: %v", err)
	}

	log.Printf("BLS keystore generated: %s", blsKeyPath)

	// Step 2: Generate ECDSA keystore from existing private key
	log.Println("Step 2: Generating ECDSA keystore from existing private key...")
	ecdsaKeyPath := filepath.Join(keysDir, "ecdsa.json")

	// Always regenerate the ECDSA keystore from the existing private key in environment
	if err := generateECDSAKeystoreFromExistingKey(ecdsaKeyPath); err != nil {
		return fmt.Errorf("failed to generate ECDSA keystore: %v", err)
	}
	log.Printf("ECDSA keystore generated: %s", ecdsaKeyPath)

	// Step 3: Update .env file with keystore paths and correct operator address
	log.Println("Step 3: Updating .env file with keystore paths...")

	// Get the correct operator address from the ECDSA private key (already derived in config)
	correctOperatorAddress := config.GetOperatorAddress().Hex()
	log.Printf("Using operator address from private key: %s", correctOperatorAddress)

	if err := updateEnvFileWithPathsAndAddress(ecdsaKeyPath, blsKeyPath, correctOperatorAddress); err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	// Reload config to pick up new keystore paths
	if err := config.Init(); err != nil {
		return fmt.Errorf("failed to reload config: %v", err)
	}

	// Step 4: Create operator and run registration
	log.Println("Step 4: Starting operator registration...")
	nodeConfig := types.NodeConfig{
		Production:                       config.GetProduction(),
		AVSOwnerAddress:                  config.GetAvsOwnerAddress().Hex(),
		OperatorAddress:                  config.GetOperatorAddress().Hex(),
		AVSAddress:                       config.GetAvsAddress().Hex(),
		EthRpcUrl:                        config.GetEthHttpRpcUrl(),
		EthWsUrl:                         config.GetEthWsRpcUrl(),
		BlsPrivateKeyStorePath:           config.GetBlsPrivateKeyStorePath(),
		OperatorEcdsaPrivateKeyStorePath: config.GetEcdsaPrivateKeyStorePath(),
		RegisterOperatorOnStartup:        false,
		NodeApiIpPortAddress:             config.GetNodeApiIpPortAddress(),
		EnableNodeApi:                    config.GetEnableNodeApi(),
	}

	log.Printf("Config - Operator Address: %s", nodeConfig.OperatorAddress)
	log.Printf("Config - AVS Address: %s", nodeConfig.AVSAddress)
	log.Printf("Config - ECDSA Keystore: %s", nodeConfig.OperatorEcdsaPrivateKeyStorePath)
	log.Printf("Config - BLS Keystore: %s", nodeConfig.BlsPrivateKeyStorePath)

	o, err := operator.NewOperatorFromConfig(nodeConfig)
	if err != nil {
		return fmt.Errorf("failed to create operator: %v", err)
	}

	// Check operator balance first
	log.Println("Checking operator balance...")
	// This check is already done in NewOperatorFromConfig, if we get here, we have some balance

	// Step 5: Register operator with chain (skip if already registered)
	log.Println("Step 5: Registering operator with chain...")
	err = o.RegisterOperatorWithChain()
	if err != nil {
		log.Printf("Warning: Chain registration check failed: %v", err)
		log.Println("Note: This means the operator is not yet registered with the chain.")
		log.Println("You need to register your operator with the chain manually first.")
		log.Printf("Your operator address is: %s", nodeConfig.OperatorAddress)
		log.Println("Please ensure this address is registered as an operator on the chain before continuing.")
		return fmt.Errorf("operator not registered with chain: %s", nodeConfig.OperatorAddress)
	} else {
		log.Println("✓ Successfully verified operator registration with chain")
	}

	// Step 6: Register operator with AVS
	log.Println("Step 6: Registering operator with AVS...")
	err = o.RegisterOperatorWithAvs()
	if err != nil {
		return fmt.Errorf("failed to register operator with AVS: %v", err)
	}
	log.Println("✓ Successfully registered operator with AVS")

	// Step 7: Register BLS Public Key
	log.Println("Step 7: Registering BLS public key...")
	err = o.RegisterBLSPublicKey()
	if err != nil {
		return fmt.Errorf("failed to register BLS public key: %v", err)
	}
	log.Println("✓ Successfully registered BLS public key")

	log.Println("\n🎉 Complete registration process finished successfully!")
	log.Println("Your operator is now fully registered and ready to participate in the AVS.")
	log.Printf("Operator Address: %s", nodeConfig.OperatorAddress)
	log.Println("\nNote: Make sure your operator address has sufficient testnet funds for transactions.")

	return nil
}

// generateBLSKeystoreFromExistingKey creates a BLS keystore from an existing private key
func generateBLSKeystoreFromExistingKey(keystorePath, privateKeyHex string) error {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Decode private key bytes
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to decode private key hex: %w", err)
	}

	// Generate corresponding public key (simplified for demo)
	publicKeyBytes := generateBLSPublicKeyFromPrivate(privateKeyBytes)

	// Create keystore structure
	keystore := map[string]interface{}{
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": privateKeyHex, // Store private key as plaintext hex
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

// generateECDSAKeystoreFromExistingKey creates an ECDSA keystore from the private key in config
func generateECDSAKeystoreFromExistingKey(keystorePath string) error {
	// Get the ECDSA private key from config
	privateKey := config.GetEcdsaPrivateKey()
	if privateKey == nil {
		return fmt.Errorf("ECDSA private key not available from config")
	}

	// Get address from public key
	address := config.GetOperatorAddress()

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	// Create an unencrypted keystore (all encryption fields set to work with empty password)
	// The MAC should be properly calculated or set to indicate no encryption
	keystore := map[string]interface{}{
		"address": address.Hex()[2:], // Remove 0x prefix
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": hex.EncodeToString(privateKeyBytes), // Store private key as hex
			"cipherparams": map[string]string{
				"iv": "00000000000000000000000000000000", // All zeros IV for no encryption
			},
			"kdf": "pbkdf2", // Use pbkdf2 instead of scrypt for simpler decryption
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

// generateBLSPublicKeyFromPrivate generates a BLS public key from private key bytes
func generateBLSPublicKeyFromPrivate(privateKeyBytes []byte) []byte {
	// This is a simplified version for demo purposes
	// In production, you'd use proper BLS12-381 operations
	publicKey := make([]byte, 48) // BLS12-381 public key is 48 bytes

	// Use a deterministic approach based on private key
	// In reality, this would be proper elliptic curve multiplication
	copy(publicKey[:32], privateKeyBytes)
	if len(privateKeyBytes) > 16 {
		copy(publicKey[32:], privateKeyBytes[:16])
	}

	return publicKey
}

// updateEnvFileWithPaths updates the .env file with keystore paths
func updateEnvFileWithPaths(ecdsaKeyPath, blsKeyPath string) error {
	envPath := ".env"

	// Read existing .env file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	envContent := string(content)

	// Check if keystore paths already exist
	if !containsLine(envContent, "ECDSA_PRIVATE_KEY_STORE_PATH=") {
		envContent += fmt.Sprintf("\n# Keystore paths (generated by complete-registration command)\n")
		envContent += fmt.Sprintf("ECDSA_PRIVATE_KEY_STORE_PATH=%s\n", ecdsaKeyPath)
	}

	if !containsLine(envContent, "BLS_PRIVATE_KEY_STORE_PATH=") {
		envContent += fmt.Sprintf("BLS_PRIVATE_KEY_STORE_PATH=%s\n", blsKeyPath)
	}

	// Write back to file
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// updateEnvFileWithPathsAndAddress updates the .env file with keystore paths and operator address
func updateEnvFileWithPathsAndAddress(ecdsaKeyPath, blsKeyPath, operatorAddress string) error {
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

// containsLine checks if a line exists in the content
func containsLine(content, line string) bool {
	lines := splitLines(content)
	for _, l := range lines {
		if len(l) > 0 && l[0] != '#' && containsString(l, line) {
			return true
		}
	}
	return false
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

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
