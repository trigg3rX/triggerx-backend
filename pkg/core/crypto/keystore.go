package crypto

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-keeper/pkg/logger"
	"go.uber.org/zap"
)

var (
	// ErrInvalidCurve represents an unsupported cryptographic curve
	ErrInvalidCurve = errors.New("invalid cryptographic curve")
	
	// ErrInvalidPassword represents password validation failures
	ErrInvalidPassword = errors.New("invalid password")
	
	// ErrKeyFileExists prevents accidental overwrite of existing keystore
	ErrKeyFileExists = errors.New("keystore file already exists")
)

type CryptoCurve string

const (
	ECDSA CryptoCurve = "ECDSA"
)

// The encryptedBLSKey struct is used to store the encrypted BLS key.
// For compatibility with the Eigenlayer keystore, use the same struct.
// https://github.com/Layr-Labs/eigensdk-go/blob/master/crypto/bls/attestation.go
type encryptedBLSKey struct {
	PubKey string              `json:"pubKey"`
	Crypto keystore.CryptoJSON `json:"crypto"`
}

func LoadPrivateKey(curve CryptoCurve, password, filePath string) ([]byte, error) {
	// Validate file existence
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.ErrorWithFields("Keystore file not found", 
			zap.String("curve", string(curve)),
			zap.String("filePath", filePath),
		)
		return nil, fmt.Errorf("keystore file not found: %s", filePath)
	}

	switch curve {
	case CryptoCurve(BLS12381), CryptoCurve(BN254):
		privKey, err := loadBLSPrivateKey(password, filePath)
		if err != nil {
			logger.ErrorWithFields("Failed to load BLS private key", 
				zap.String("curve", string(curve)),
				zap.Error(err),
			)
			return nil, err
		}
		return privKey, nil
	case "ECDSA":
		pk, err := loadECDSAPrivateKey(password, filePath)
		if err != nil {
			logger.ErrorWithFields("Failed to load ECDSA private key", 
				zap.Error(err),
			)
			return nil, err
		}
		return ecrypto.FromECDSA(pk), nil
	default:
		logger.ErrorWithFields("Unsupported curve for key loading", 
			zap.String("curve", string(curve)),
		)
		return nil, ErrInvalidCurve
	}
}

// ReadKeystorePasswordFromFile reads the password from the password file.
func ReadKeystorePasswordFromFile(passwordFilePath string) (string, error) {
	password, err := os.ReadFile(passwordFilePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(password)), nil
}

func SaveKey(curve CryptoCurve, privKey []byte, password, filePath string) error {
	// Validate password
	if err := validatePassword(password); err != nil {
		logger.ErrorWithFields("Password validation failed", 
			zap.String("curve", string(curve)),
			zap.Error(err),
		)
		return err
	}

	// Check file availability
	if err := checkKeystoreFileAvailability(filePath); err != nil {
		return err
	}

	switch curve {
	case CryptoCurve(BLS12381), CryptoCurve(BN254):
		return saveBLSKey(BLSCurve(curve), privKey, password, filePath)
	case "ECDSA":
		return saveECDSAKey(privKey, password, filePath)
	default:
		logger.ErrorWithFields("Unsupported curve for key generation", 
			zap.String("curve", string(curve)),
		)
		return ErrInvalidCurve
	}
}

func saveBLSKey(curve BLSCurve, privKey []byte, password, filePath string) error {
	blsScheme := NewBLSScheme(curve)
	pubKey, err := blsScheme.GetPublicKey(privKey, false, true)
	if err != nil {
		return err
	}

	encryptedKey, err := keystore.EncryptDataV3(privKey, []byte(password), keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	encKey := encryptedBLSKey{
		PubKey: common.Bytes2Hex(pubKey),
		Crypto: encryptedKey,
	}
	encKeyData, err := json.Marshal(encKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, encKeyData, 0644)
}

func loadBLSPrivateKey(password, filePath string) ([]byte, error) {
	ksData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	encBLSStruct := &encryptedBLSKey{}
	if err := json.Unmarshal(ksData, encBLSStruct); err != nil {
		return nil, err
	}

	if encBLSStruct.PubKey == "" {
		return nil, errors.New("invalid bls key file, missing public key")
	}

	return keystore.DecryptDataV3(encBLSStruct.Crypto, password)
}

func saveECDSAKey(privKey []byte, password, filePath string) error {
	privateKey, err := ecrypto.ToECDSA(privKey)
	if err != nil {
		return err
	}

	UUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	encKey := &keystore.Key{
		Id:         UUID,
		Address:    ecrypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}
	encKeyData, err := keystore.EncryptKey(encKey, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, encKeyData, 0644)
}

func loadECDSAPrivateKey(password, filePath string) (*ecdsa.PrivateKey, error) {
	ksData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	key, err := keystore.DecryptKey(ksData, password)
	if err != nil {
		return nil, err
	}

	return key.PrivateKey, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrInvalidPassword)
	}
	
	// Check for at least one uppercase, one lowercase, one number, and one special character
	if match, _ := regexp.MatchString(`^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()]).{8,}$`, password); !match {
		return fmt.Errorf("%w: password must contain uppercase, lowercase, number, and special character", ErrInvalidPassword)
	}
	
	return nil
}

func checkKeystoreFileAvailability(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		logger.ErrorWithFields("Keystore file already exists",
			zap.String("filePath", filePath),
		)
		return fmt.Errorf("%w: file %s already exists", ErrKeyFileExists, filePath)
	}
	return nil
}

func ListKeystoreFiles(directory string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(directory, "*"))
	if err != nil {
		logger.ErrorWithFields("Failed to list keystore files", 
			zap.String("directory", directory),
			zap.Error(err),
		)
		return nil, err
	}
	
	var keystoreFiles []string
	for _, file := range files {
		if filepath.Ext(file) == ".json" {
			keystoreFiles = append(keystoreFiles, file)
		}
	}
	
	return keystoreFiles, nil
}