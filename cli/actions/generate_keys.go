package actions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/trigg3rX/triggerx-backend/cli/core"
	"github.com/urfave/cli"
)

func GenerateKeys(ctx *cli.Context) error {
	log.Println("Generating operator keys...")

	// Create keys directory
	keysDir := "keys"
	if err := os.MkdirAll(keysDir, 0755); err != nil {
		return fmt.Errorf("failed to create keys directory: %v", err)
	}

	// Generate ECDSA keystore
	log.Println("Generating ECDSA keystore...")
	ecdsaKeyPath := filepath.Join(keysDir, "ecdsa.json")
	if err := core.GenerateECDSAKeystore(ecdsaKeyPath); err != nil {
		return fmt.Errorf("failed to generate ECDSA keystore: %v", err)
	}

	// Generate BLS keystore
	log.Println("Generating BLS keystore...")
	blsKeyPath := filepath.Join(keysDir, "bls.json")
	blsPrivateKeyHex, err := core.GenerateBLSKeystore(blsKeyPath)
	if err != nil {
		return fmt.Errorf("failed to generate BLS keystore: %v", err)
	}

	// Update .env file with keystore paths
	envPath := ".env"
	if err := core.UpdateEnvFile(envPath, ecdsaKeyPath, blsKeyPath, blsPrivateKeyHex); err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	log.Println("Key generation completed successfully!")
	log.Printf("ECDSA keystore: %s", ecdsaKeyPath)
	log.Printf("BLS keystore: %s", blsKeyPath)
	log.Printf("Updated .env file with keystore paths")
	return nil
}
