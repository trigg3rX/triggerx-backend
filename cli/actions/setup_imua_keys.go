package actions

import (
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

	log.Printf("Setup configuration:")
	log.Printf("- Home Directory: %s", imuaHomeDir)
	log.Printf("- Account Key Name: %s", imuaAccountKeyName)

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
	keyExists, err := checkValidatorKeyExists(imuaHomeDir, imuaAccountKeyName)
	if err != nil {
		return fmt.Errorf("failed to check if validator key exists: %v", err)
	}

	if keyExists {
		log.Printf("✅ Validator key '%s' already exists", imuaAccountKeyName)

		// Show the address
		getAddrCmd := exec.Command("imuad",
			"--home", imuaHomeDir,
			"keys", "show", "-a", imuaAccountKeyName,
			"--keyring-backend", "test",
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

	// Step 4: Create validator key
	// log.Printf("\n🚀 Step 4: Creating validator key '%s'...", imuaAccountKeyName)
	// if err := createValidatorKey(imuaHomeDir, imuaAccountKeyName); err != nil {
	// 	return fmt.Errorf("failed to create validator key: %v", err)
	// }

	// log.Println("\n🎉 Imuachain key setup completed successfully!")
	// log.Println("\n📋 Summary:")
	// log.Printf("   ✅ Validator key '%s' created", imuaAccountKeyName)
	// log.Printf("   ✅ Home directory: %s", imuaHomeDir)
	// log.Println("\n💡 Next Steps:")
	// log.Println("   1. Fund your validator address with IMUA tokens from the testnet faucet")
	// log.Println("   2. Run 'triggerx register-imua-operator' to register as an operator")
	// log.Println("   3. Run 'triggerx complete-imua-registration' for full setup")

	return nil
}

// ensureImuaHomeDir creates the imua home directory if it doesn't exist
func ensureImuaHomeDir(imuaHomeDir string) error {
	// Check if directory exists
	if _, err := os.Stat(imuaHomeDir); os.IsNotExist(err) {
		log.Printf("Creating home directory: %s", imuaHomeDir)

		// Initialize imuad configuration
		initCmd := exec.Command("imuad", "init", "validator", "--home", imuaHomeDir)
		output, err := initCmd.CombinedOutput()
		if err != nil {
			log.Printf("Init command output: %s", string(output))
			return fmt.Errorf("failed to initialize imuad home directory: %v", err)
		}

		log.Printf("Initialized imuad home directory")
	} else if err != nil {
		return fmt.Errorf("failed to check home directory: %v", err)
	}

	return nil
}

// checkValidatorKeyExists checks if the validator key already exists
func checkValidatorKeyExists(imuaHomeDir, keyName string) (bool, error) {
	listCmd := exec.Command("imuad",
		"--home", imuaHomeDir,
		"keys", "list",
		"--output", "json",
		"--keyring-backend", "test",
	)

	output, err := listCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list keys: %v", err)
	}

	outputStr := string(output)
	// Simple check - look for the key name in the JSON output
	// In production, you'd want proper JSON parsing
	return strings.Contains(outputStr, fmt.Sprintf(`"name":"%s"`, keyName)), nil
}

// createValidatorKey creates a new validator key
// func createValidatorKey(imuaHomeDir, keyName string) error {
// 	createCmd := exec.Command("imuad",
// 		"--home", imuaHomeDir,
// 		"keys", "add", keyName, "--keyring-backend", "test",
// 	)

// 	output, err := createCmd.CombinedOutput()
// 	if err != nil {
// 		log.Printf("Create key command output: %s", string(output))
// 		return fmt.Errorf("failed to create validator key: %v", err)
// 	}

// 	log.Printf("Validator key created successfully:")
// 	log.Printf("%s", string(output))

// 	return nil
// }

// EnsureValidatorKeyExists is a helper function that can be called by other commands
// to ensure the validator key exists before proceeding
func EnsureValidatorKeyExists() error {
	imuaHomeDir := os.Getenv("IMUA_HOME_DIR")
	if imuaHomeDir == "" {
		return fmt.Errorf("IMUA_HOME_DIR environment variable not set")
	}

	imuaAccountKeyName := os.Getenv("IMUA_ACCOUNT_KEY_NAME")
	if imuaAccountKeyName == "" {
		return fmt.Errorf("IMUA_ACCOUNT_KEY_NAME environment variable not set")
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
	keyExists, err := checkValidatorKeyExists(imuaHomeDir, imuaAccountKeyName)
	if err != nil {
		return fmt.Errorf("failed to check if validator key exists: %v", err)
	}

	if !keyExists {
		log.Printf("⚠️ Validator key '%s' not found", imuaAccountKeyName)
		// log.Println("Creating validator key automatically...")

		// if err := createValidatorKey(imuaHomeDir, imuaAccountKeyName); err != nil {
		// 	return fmt.Errorf("failed to create validator key: %v", err)
		// }

		// log.Printf("✅ Validator key '%s' created successfully", imuaAccountKeyName)
	}

	return nil
}
