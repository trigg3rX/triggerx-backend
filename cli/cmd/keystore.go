package cmd

// import (
// 	"fmt"
// 	"os"
// 	"path/filepath"

// 	ecrypto "github.com/ethereum/go-ethereum/crypto"

// 	"github.com/spf13/cobra"

// 	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
// 	"github.com/Layr-Labs/eigensdk-go/crypto/ecdsa"
// 	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

// 	"github.com/Layr-Labs/eigensdk-go/logging"

// 	"github.com/trigg3rX/triggerx-avs-cli/utils"
// )

// var generateKeystoreCmd = &cobra.Command{
// 	Use:   "generate-keystore",
// 	Short: "Generate keystore",
// 	Long: `
// 	Generate a new keystore for the operator; user can choose between ECDSA and BLS key types.
// 	The input required is a passphrase for the new keystore. 
// 	Passphrase must contain:
// 		- At least 8 characters
// 		- At least one uppercase letter
// 		- At least one lowercase letter
// 		- At least one number
// 		- At least one special character
// 	The keystore file will be encrypted with the provided passphrase. The same passphrase will be used to decrypt the keystore file when signing transactions.
// 	`,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		generateKeystore()
// 	},
// }

// var logger logging.Logger

// func init() {
// 	var err error
// 	logger, err = logging.NewZapLogger(logging.Development)
// 	if err != nil {
// 		logger.Fatalf("Failed to initialize logger: %v", err)
// 	}
// }

// func generateKeystore() {
// 	fmt.Println(`
// 	Generate a new keystore for the operator.
// 	The input required is a passphrase for the new keystore. 
// 	Passphrase must contain:
// 		- At least 8 characters
// 		- At least one uppercase letter
// 		- At least one lowercase letter
// 		- At least one number
// 		- At least one special character
// 	The keystore file will be encrypted with the provided passphrase. The same passphrase will be used to decrypt the keystore file when signing transactions.
// 	Keep the passphrase safe as it is required to decrypt the keystore file when signing transactions.
// 	`)
// 	logger.Info("Starting keystore generation process")
// 	err := generateAndSaveKeystore()
// 	if err != nil {
// 		logger.Fatalf("Failed to generate and save keystore: %v", err)
// 	}

// 	logger.Info("Keystore generation completed successfully")
// }

// func generateAndSaveKeystore() error {
// 	passphrase, err := utils.PasswordPrompt("Enter passphrase for the new keystore: ", true)
// 	if err != nil {
// 		logger.Errorf("Failed to get passphrase: %v", err)
// 		return err
// 	}

// 	// Generate ECDSA and BLS keys
// 	privateKey, err := ecrypto.GenerateKey()
// 	if err != nil {
// 		logger.Errorf("Failed to generate ECDSA key: %v", err)
// 		return err
// 	}
// 	eprivKey := ecrypto.FromECDSA(privateKey)
// 	addr := ecrypto.PubkeyToAddress(privateKey.PublicKey)
// 	err = saveKeystore("ecdsa", passphrase, addr.Bytes(), eprivKey)
// 	if err != nil {
// 		return err
// 	}

// 	keyPair, err := bls.GenRandomBlsKeys()
// 	if err != nil {
// 		logger.Errorf("Failed to generate BLS key: %v", err)
// 		return err
// 	}
// 	pubKey := keyPair.GetPubKeyG1().Serialize()
// 	bytes := keyPair.PrivKey.Bytes()
// 	privKey := bytes[:]
// 	return saveKeystore("bls", passphrase, pubKey, privKey)


// 	// keyType, err := utils.StringPrompt("Enter key type (ecdsa/bls): ", false, true, []string{"ecdsa", "bls"})
// 	// if err != nil {
// 	// 	logger.Errorf("Failed to get key type: %v", err)
// 	// 	return err
// 	// }

// 	// keyType = strings.ToLower(strings.TrimSpace(keyType))
// 	// if keyType != "ecdsa" && keyType != "bls" {
// 	// 	logger.Fatalf("Invalid key type. Must be either 'ecdsa' or 'bls'")
// 	// }

// 	// logger.Info("Generating keys", "keyType", keyType)

// 	// switch keyType {
// 	// case "ecdsa":
// 	// 	privateKey, err := ecrypto.GenerateKey()
// 	// 	if err != nil {
// 	// 		logger.Errorf("Failed to generate ECDSA key: %v", err)
// 	// 		return err
// 	// 	}
// 	// 	privKey := ecrypto.FromECDSA(privateKey)
// 	// 	addr := ecrypto.PubkeyToAddress(privateKey.PublicKey)
// 	// 	return saveKeystore(keyType, passphrase, addr.Bytes(), privKey)
// 	// case "bls":
// 	// 	keyPair, err := bls.GenRandomBlsKeys()
// 	// 	if err != nil {
// 	// 		logger.Errorf("Failed to generate BLS key: %v", err)
// 	// 		return err
// 	// 	}
// 	// 	pubKey := keyPair.GetPubKeyG1().Serialize()
// 	// 	privKey := keyPair.PrivKey.Bytes()
// 	// 	return saveKeystore(keyType, passphrase, pubKey, privKey[:])
// 	// default:
// 	// 	logger.Fatalf("Unhandled key type: %s", keyType)
// 	// 	return nil
// 	// }
// }

// func saveKeystore(keyType string, passphrase string, pubKey, privKey []byte) error {
// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		logger.Errorf("Failed to get user home directory: %v", err)
// 		return err
// 	}
// 	ksPath := filepath.Join(homeDir, ".triggerx/keystore", fmt.Sprintf("%s_%x.key", keyType, pubKey[:6]))

// 	if err := utils.DisplayWarningMessage(keyType, utils.Bytes2Hex(privKey), ksPath); err != nil {
// 		logger.Errorf("Failed to display warning message: %v", err)
// 		return err
// 	}

// 	logger.Info("Saving keystore", "path", ksPath, "keyType", keyType)

// 	if keyType == "ecdsa" {
// 		privKeyHex := utils.Bytes2Hex(privKey)
// 		err = ecdsa.WriteKeyFromHex(ksPath, privKeyHex, passphrase)
// 		if err != nil {
// 			logger.Errorf("Failed to write ECDSA key to file: %v", err)
// 		}
// 	} else {
// 		privKeyElement := new(fr.Element)
// 		privKeyElement.SetBytes(privKey)
// 		keyPair := bls.NewKeyPair(privKeyElement)
// 		err = keyPair.SaveToFile(ksPath, passphrase)
// 		if err != nil {
// 			logger.Errorf("Failed to save BLS key to file: %v", err)
// 		}
// 	}

// 	if err != nil {
// 		return err
// 	}

// 	logger.Info("Keystore saved successfully", "path", ksPath)
// 	return nil
// }