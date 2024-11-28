package crypto

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/trigg3rX/triggerx-keeper/pkg/core/errors"
	"github.com/trigg3rX/triggerx-keeper/pkg/core/logger"
	"go.uber.org/zap"
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

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", errors.ErrInvalidPassword)
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()]`).MatchString(password)
	
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("%w: password must contain uppercase, lowercase, number, and special character", errors.ErrInvalidPassword)
	}
	
	return nil
}

func checkKeystoreFileAvailability(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		logger.ErrorWithFields("Keystore file already exists",
			zap.String("filePath", filePath),
		)
		return fmt.Errorf("%w: file %s already exists", errors.ErrKeyFileExists, filePath)
	}
	return nil
}

func LoadPrivateKey(curve CryptoCurve, password, filePath string) ([]byte, error) {
	// Validate file existence
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.ErrorWithFields("Keystore file not found", 
			zap.String("curve", string(curve)),
			zap.String("filePath", filePath),
		)
		return nil, errors.ErrKeystoreNotFound
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
		return nil, errors.ErrInvalidCurve
	}
}

func loadBLSPrivateKey(password, filePath string) ([]byte, error) {
	ksData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.ErrKeystoreNotFound
	}

	encBLSStruct := &encryptedBLSKey{}
	if err := json.Unmarshal(ksData, encBLSStruct); err != nil {
		return nil, errors.ErrKeystoreDecryption
	}

	if encBLSStruct.PubKey == "" {
		return nil, errors.ErrInvalidBLSKeyFile
	}

	privKey, err := keystore.DecryptDataV3(encBLSStruct.Crypto, password)
	if err != nil {
		return nil, errors.ErrKeystoreDecryption
	}
	return privKey, nil
}

func loadECDSAPrivateKey(password, filePath string) (*ecdsa.PrivateKey, error) {
	ksData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.ErrKeystoreNotFound
	}

	key, err := keystore.DecryptKey(ksData, password)
	if err != nil {
		return nil, errors.ErrKeystoreDecryption
	}

	return key.PrivateKey, nil
}

func SaveKey(curve CryptoCurve, privKey []byte, password, filePath string) error {
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
		return errors.ErrInvalidCurve
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
		return errors.ErrKeystoreEncryption
	}

	encKey := encryptedBLSKey{
		PubKey: common.Bytes2Hex(pubKey),
		Crypto: encryptedKey,
	}
	encKeyData, err := json.Marshal(encKey)
	if err != nil {
		return errors.ErrKeystoreEncryption
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.ErrDirectoryCreation
	}

	return os.WriteFile(filePath, encKeyData, 0644)
}

func saveECDSAKey(privKey []byte, password, filePath string) error {
	privateKey, err := ecrypto.ToECDSA(privKey)
	if err != nil {
		return errors.ErrInvalidPrivateKey
	}

	UUID, err := uuid.NewRandom()
	if err != nil {
		return errors.ErrKeystoreEncryption
	}

	encKey := &keystore.Key{
		Id:         UUID,
		Address:    ecrypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}
	encKeyData, err := keystore.EncryptKey(encKey, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return errors.ErrKeystoreEncryption
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.ErrDirectoryCreation
	}

	return os.WriteFile(filePath, encKeyData, 0644)
}

func ListKeystoreFiles(directory string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(directory, "*"))
	if err != nil {
		logger.ErrorWithFields("Failed to list keystore files", 
			zap.String("directory", directory),
			zap.Error(err),
		)
		return nil, errors.ErrKeystoreListing
	}
	
	var keystoreFiles []string
	for _, file := range files {
		if filepath.Ext(file) == ".json" {
			keystoreFiles = append(keystoreFiles, file)
		}
	}
	
	return keystoreFiles, nil
}