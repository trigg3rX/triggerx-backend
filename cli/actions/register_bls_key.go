package actions

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/imua-xyz/imua-avs-sdk/logging"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/urfave/cli"
)

// BLS12-381 curve order
const BLS12381_CURVE_ORDER = "52435875175126190479447740508185965837690552500527637822603658699938581184513"

func RegisterBLSPublicKey(c *cli.Context) error {
	logger := logging.NewSlogLogger(logging.Development)
	logger.Info("Starting BLS Public Key Registration...")

	// Load environment variables
	requiredEnvVars := map[string]string{
		"AVS_ADDRESS":      env.GetEnvString("AVS_ADDRESS", ""),
		"BLS_PRIVATE_KEY":  env.GetEnvString("BLS_PRIVATE_KEY", ""),
		"IMUA_BECH32_KEY":  env.GetEnvString("IMUA_BECH32_KEY", ""),
		"IMUA_CHAIN_ID":    env.GetEnvString("IMUA_CHAIN_ID", ""),
		"IMUA_HEX_KEY":     env.GetEnvString("IMUA_HEX_KEY", ""),
		"IMUA_PRIVATE_KEY": env.GetEnvString("IMUA_PRIVATE_KEY", ""),
		"ETH_HTTP_RPC_URL": env.GetEnvString("ETH_HTTP_RPC_URL", ""),
		"ENV_FILE_PATH":    env.GetEnvString("ENV_FILE_PATH", ".env"),
	}

	// Validate required environment variables
	for key, value := range requiredEnvVars {
		if value == "" {
			logger.Error("Required environment variable not set", "variable", key)
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	// Load and validate BLS private key
	blsPrivateKeyHex := requiredEnvVars["BLS_PRIVATE_KEY"]
	validatedKey, err := validateAndFixBLSKey(blsPrivateKeyHex)
	if err != nil {
		logger.Error("BLS key validation failed", "error", err)
		return fmt.Errorf("BLS key validation failed: %w", err)
	}

	blsPrivateKey, err := parseBLSPrivateKey(validatedKey)
	if err != nil {
		logger.Error("Failed to parse BLS private key", "error", err)
		return fmt.Errorf("failed to parse BLS private key: %w", err)
	}

	// Get BLS public key
	blsPublicKey := blsPrivateKey.PublicKey()
	blsPublicKeyBytes := blsPublicKey.Marshal()
	blsPublicKeyHex := "0x" + hex.EncodeToString(blsPublicKeyBytes)

	// Create message to sign for BLS registration
	chainID := requiredEnvVars["IMUA_CHAIN_ID"]
	operatorBech32 := requiredEnvVars["IMUA_BECH32_KEY"]
	chainIDWithoutRevision := formatChainIDWithoutRevision(chainID)

	messageToSign := fmt.Sprintf("BLS12-381 Signed Message\nChainIDWithoutRevision: %s\nAccAddressBech32: %s",
		chainIDWithoutRevision, operatorBech32)

	// Hash the message using Keccak256
	messageHash := crypto.Keccak256Hash([]byte(messageToSign))

	// Sign the message hash with BLS private key
	signature := blsPrivateKey.Sign(messageHash[:])
	signatureBytes := signature.Marshal()
	signatureHex := "0x" + hex.EncodeToString(signatureBytes)

	envFilePath := requiredEnvVars["ENV_FILE_PATH"]
	err = storeSignatureInEnvFile(envFilePath, signatureHex)
	if err != nil {
		logger.Error("Failed to store BLS signature in .env file", "error", err)
		return fmt.Errorf("failed to store BLS signature in .env file: %w", err)
	}

	// Prepare the cast command
	avsAddress := requiredEnvVars["AVS_ADDRESS"]
	operatorAddress := requiredEnvVars["IMUA_HEX_KEY"]
	privateKey := requiredEnvVars["IMUA_PRIVATE_KEY"]
	rpcUrl := requiredEnvVars["ETH_HTTP_RPC_URL"]

	castCommand := fmt.Sprintf(
		"cast send %s \"registerBLSPublicKey(address,bytes,bytes)\" %s %s %s --rpc-url %s --private-key %s --gas-limit 1000000",
		avsAddress,
		avsAddress,
		blsPublicKeyHex,
		signatureHex,
		rpcUrl,
		privateKey,
	)

	// Print the command being executed
	fmt.Println("🚀 Executing cast command:")
	fmt.Println(castCommand)
	fmt.Println("")

	// Execute the cast command
	cmd := exec.Command("bash", "-c", castCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("❌ Error executing cast command:")
		fmt.Println(string(output))
		return fmt.Errorf("failed to execute cast command: %w", err)
	}

	// Print the output
	fmt.Println("✅ BLS Public Key registration successful!")
	fmt.Println("📋 Command output:")
	fmt.Println(string(output))
	fmt.Println("\n📋 Details:")
	fmt.Printf("🔑 BLS Public Key: %s\n", blsPublicKeyHex)
	fmt.Printf("🏢 Operator Address: %s\n", operatorAddress)
	fmt.Printf("📝 Message Signed: %s\n", messageToSign)
	fmt.Printf("🔐 Signature: %s\n", signatureHex)
	fmt.Printf("🌐 RPC URL: %s\n", rpcUrl)

	return nil
}

func parseBLSPrivateKey(hexKey string) (bls.SecretKey, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")

	keyBytes := common.FromHex(hexKey)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("BLS private key must be 32 bytes, got %d", len(keyBytes))
	}

	curveOrder := new(big.Int)
	curveOrder.SetString(BLS12381_CURVE_ORDER, 10)

	keyInt := new(big.Int).SetBytes(keyBytes)
	if keyInt.Cmp(curveOrder) >= 0 {
		return nil, fmt.Errorf("private key is too large for BLS12-381 curve")
	}

	if keyInt.Sign() == 0 {
		return nil, fmt.Errorf("private key cannot be zero")
	}

	return bls.SecretKeyFromBytes(keyBytes)
}

func validateAndFixBLSKey(currentKey string) (string, error) {
	currentKey = strings.TrimPrefix(currentKey, "0x")

	if len(currentKey) != 64 {
		return "", fmt.Errorf("BLS key must be 64 hex characters, got %d", len(currentKey))
	}

	keyBytes := common.FromHex(currentKey)
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("invalid hex encoding")
	}

	curveOrder := new(big.Int)
	curveOrder.SetString(BLS12381_CURVE_ORDER, 10)

	keyInt := new(big.Int).SetBytes(keyBytes)

	if keyInt.Cmp(curveOrder) >= 0 {
		reducedKey := new(big.Int).Mod(keyInt, curveOrder)
		if reducedKey.Sign() == 0 {
			reducedKey.SetInt64(1)
		}

		reducedBytes := make([]byte, 32)
		reducedKeyBytes := reducedKey.Bytes()
		copy(reducedBytes[32-len(reducedKeyBytes):], reducedKeyBytes)

		return hex.EncodeToString(reducedBytes), nil
	}

	return currentKey, nil
}

func formatChainIDWithoutRevision(chainID string) string {
	lastDash := strings.LastIndex(chainID, "-")
	if lastDash == -1 {
		return chainID
	}

	revision := chainID[lastDash+1:]
	if len(revision) > 0 {
		for _, r := range revision {
			if r < '0' || r > '9' {
				return chainID
			}
		}
		return chainID[:lastDash]
	}

	return chainID
}
func storeSignatureInEnvFile(filePath, signature string) error {
	// Open the file in append mode, create if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close .env file: %v\n", cerr)
		}
	}()

	// Check if BLS_SIGNATURE already exists in the file
	var hasSignature bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "BLS_SIGNATURE=") {
			hasSignature = true
			break
		}
	}

	// If we found an existing signature, we should update it rather than append
	if hasSignature {
		// Read the entire file
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to beginning of .env file: %w", err)
		}
		var lines []string
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "BLS_SIGNATURE=") {
				line = "BLS_SIGNATURE=" + signature
			}
			lines = append(lines, line)
		}

		// Truncate the file and rewrite all lines
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close .env file: %w", err)
		}
		file, err = os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to truncate .env file: %w", err)
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				fmt.Printf("Warning: failed to close .env file: %v\n", cerr)
			}
		}()

		writer := bufio.NewWriter(file)
		for _, line := range lines {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to .env file: %w", err)
			}
		}
		return writer.Flush()
	}

	// If no existing signature, just append it
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("BLS_SIGNATURE=" + signature + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to .env file: %w", err)
	}
	return writer.Flush()
}
