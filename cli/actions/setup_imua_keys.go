package actions

import (
	// "encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/trigg3rX/triggerx-backend/cli/core/config"
	"github.com/urfave/cli"
)

func SetupImuaKeys(ctx *cli.Context) error {
	log.Println("🔑 Setting up Imuachain keys...")

	// Initialize config from environment variables
	err := config.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	// Check required environment variables
	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
	}

	// Get keyring backend from env or use default
	keyringBackend := os.Getenv("IMUA_KEYRING_BACKEND")
	if keyringBackend == "" {
		keyringBackend = "file" // default to file backend
	}

	log.Printf("Setup configuration:")
	log.Printf("- Home Directory: %s", imuaHomeDir)
	log.Printf("- Account Key Name: %s", imuaAccountKeyName)
	log.Printf("- Keyring Backend: %s", keyringBackend)

	// Step 1: Check if imuad is available
	log.Println("\n🔍 Step 1: Checking if imuad is available...")
	_, err = exec.LookPath("imuad")
	if err != nil {
		return fmt.Errorf("imuad binary not found in PATH. Please install imuad first")
	}
	log.Println("✅ imuad binary found")

	// Step 2: Ensure home directory exists
	log.Println("\n📁 Step 2: Ensuring home directory exists...")
	if err := ensureImuaHomeDir(imuaHomeDir); err != nil {
		return fmt.Errorf("failed to setup home directory: %v", err)
	}
	log.Printf("✅ Home directory ready: %s", imuaHomeDir)

	// Step 3: Check if validator key exists
	log.Println("\n🔑 Step 3: Checking if validator key exists...")
	keyExists, err := checkValidatorKeyExists(imuaHomeDir, imuaAccountKeyName, keyringBackend)
	if err != nil {
		return fmt.Errorf("failed to check if validator key exists: %v", err)
	}

	if keyExists {
		log.Printf("✅ Validator key '%s' already exists", imuaAccountKeyName)

		// Show the address
		getAddrCmd := exec.Command("imuad",
			"--home", imuaHomeDir,
			"keys", "show", "-a", imuaAccountKeyName,
			"--keyring-backend", keyringBackend,
		)

		addrOutput, err := getAddrCmd.Output()
		if err != nil {
			log.Printf("⚠️ Could not retrieve address: %v", err)
		} else {
			address := strings.TrimSpace(string(addrOutput))
			log.Printf("   Address: %s", address)
		}

		return nil
	}

	// Step 4: Create the key if it doesn't exist
	log.Printf("\n🚀 Step 4: Creating validator key '%s'...", imuaAccountKeyName)
	if err := createValidatorKey(imuaHomeDir, imuaAccountKeyName, keyringBackend); err != nil {
		return fmt.Errorf("failed to create validator key: %v", err)
	}

	log.Println("\n🎉 Imuachain key setup completed successfully!")
	log.Println("\n📋 Summary:")
	log.Printf("   ✅ Validator key '%s' created", imuaAccountKeyName)
	log.Printf("   ✅ Home directory: %s", imuaHomeDir)
	log.Printf("   ✅ Keyring backend: %s", keyringBackend)

	return nil
}

// ensureImuaHomeDir creates the imua home directory if it doesn't exist
func ensureImuaHomeDir(imuaHomeDir string) error {
	// Check if directory exists
	if _, err := os.Stat(imuaHomeDir); os.IsNotExist(err) {
		log.Printf("Creating home directory: %s", imuaHomeDir)
		err := os.MkdirAll(imuaHomeDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create home directory: %v", err)
		}
		log.Printf("Created imuad home directory")
	} else if err != nil {
		return fmt.Errorf("failed to check home directory: %v", err)
	}

	return nil
}

// checkValidatorKeyExists checks if the validator key already exists
// checkValidatorKeyExists checks if the validator key already exists
func checkValidatorKeyExists(imuaHomeDir, keyName, keyringBackend string) (bool, error) {
	listCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "list",
		"--output", "json",
		"--keyring-backend", keyringBackend,
	)

	// For file backend, we need to provide the password
	if keyringBackend == "file" {
		keyringPassword := os.Getenv("KEYRING_PASSWORD")
		if keyringPassword != "" {
			listCmd.Stdin = strings.NewReader(keyringPassword + "\n")
		}
	}

	output, err := listCmd.Output()
	if err != nil {
		// If the command fails, it might be because no keys exist yet
		// Check if the error is about no keys being available
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no keys found") || strings.Contains(stderr, "keyring is empty") {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to list keys: %v", err)
	}

	outputStr := string(output)
	// Simple check - look for the key name in the JSON output
	return strings.Contains(outputStr, fmt.Sprintf(`"name":"%s"`, keyName)), nil
}

// createValidatorKey creates a new validator key
func createValidatorKey(imuaHomeDir, keyName, keyringBackend string) error {
	createCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "add", keyName,
		"--keyring-backend", keyringBackend,
	)

	// For file backend, we might need to handle password input
	// For now, let's assume it will prompt the user if needed
	createCmd.Stdin = os.Stdin
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	log.Printf("Creating key '%s' with %s backend...", keyName, keyringBackend)

	if keyringBackend == "file" {
		log.Println("⚠️  File backend requires a password. You will be prompted to enter and confirm a password.")
	}

	err := createCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create validator key: %v", err)
	}

	log.Printf("✅ Validator key '%s' created successfully", keyName)
	return nil
}

// EnsureValidatorKeyExists is a helper function that can be called by other commands
// to ensure the validator key exists before proceeding
// Replace your existing EnsureValidatorKeyExists function with this:
func EnsureValidatorKeyExists() error {
	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
	}

	// Fix: Use IMUA_KEYRING_BACKEND instead of KEYRING_BACKEND
	keyringBackend := os.Getenv("KEYRING_BACKEND")
	if keyringBackend == "" {
		keyringBackend = "file"
	}

	// Check if imuad is available
	_, err := exec.LookPath("imuad")
	if err != nil {
		return fmt.Errorf("imuad binary not found in PATH. Please install imuad first")
	}

	// Ensure home directory exists
	if err := ensureImuaHomeDir(imuaHomeDir); err != nil {
		return fmt.Errorf("failed to setup home directory: %v", err)
	}

	// Check if key exists
	keyExists, err := checkValidatorKeyExists(imuaHomeDir, imuaAccountKeyName, keyringBackend)
	if err != nil {
		return fmt.Errorf("failed to check if validator key exists: %v", err)
	}

	if !keyExists {
		log.Printf("⚠️ Validator key '%s' not found, creating it...", imuaAccountKeyName)
		if err := createValidatorKey(imuaHomeDir, imuaAccountKeyName, keyringBackend); err != nil {
			return fmt.Errorf("failed to create validator key: %v", err)
		}
		log.Printf("✅ Validator key '%s' created successfully", imuaAccountKeyName)
	} else {
		log.Printf("✅ Validator key '%s' already exists", imuaAccountKeyName)
	}

	return nil
}
