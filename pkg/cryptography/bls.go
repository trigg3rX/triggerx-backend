package cryptography

// import (
// 	"encoding/json"
// 	"fmt"
// 	"math/big"
// 	"strings"

// 	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
// 	"github.com/ethereum/go-ethereum/crypto"
// )

// // SignBLSMessage signs a message using BLS signature scheme
// func SignBLSMessage(message []byte, privateKey string) (string, error) {
// 	// Create BLS key pair from private key string
// 	decimalKey, err := HexToBLSPrivateKey(privateKey)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid private key: %w", err)
// 	}
// 	keyPair, err := bls.NewKeyPairFromString(decimalKey)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid private key: %w", err)
// 	}

// 	fmt.Println("private key", keyPair.PrivKey.String())
// 	fmt.Println("public key", keyPair.GetPubKeyG2().String())

// 	// Hash to 32-byte array for BLS signing
// 	var messageArray [32]byte
// 	copy(messageArray[:], message[:])

// 	// Sign the message using BLS
// 	signature := keyPair.SignMessage(messageArray)

// 	// Format signature as JSON with hex-encoded x and y coordinates
// 	// This matches the format in mcl.js g1ToHex() and sign() functions
// 	xBytes := signature.G1Point.X.BigInt(new(big.Int)).Bytes()
// 	yBytes := signature.G1Point.Y.BigInt(new(big.Int)).Bytes()

// 	// Format hex values with 0x prefix and proper padding
// 	xHex := fmt.Sprintf("0x%064x", new(big.Int).SetBytes(xBytes))
// 	yHex := fmt.Sprintf("0x%064x", new(big.Int).SetBytes(yBytes))

// 	jsonSig := map[string]string{
// 		"x": xHex,
// 		"y": yHex,
// 	}

// 	jsonBytes, err := json.Marshal(jsonSig)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal signature: %w", err)
// 	}
// 	return string(jsonBytes), nil
// }

// // SignBLSJSONMessage signs JSON data using BLS
// func SignBLSJSONMessage(jsonData interface{}, privateKey string) (string, error) {
// 	jsonDataMap := make(map[string]interface{})

// 	jsonBytes, err := json.Marshal(jsonData)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal input data: %w", err)
// 	}

// 	if err := json.Unmarshal(jsonBytes, &jsonDataMap); err != nil {
// 		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
// 	}

// 	// Convert to lowercase (same as ECDSA implementation)
// 	convertToLower(jsonDataMap)

// 	jsonDataBytes, err := json.Marshal(jsonDataMap)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal json data: %w", err)
// 	}

// 	return SignBLSMessage(jsonDataBytes, privateKey)
// }

// func VerifyBLSSignature(message string, signature string, pubkeyData string) (bool, error) {
// 	// Parse signature
// 	sig, err := deserializeBLSSignature(signature)
// 	if err != nil {
// 		return false, fmt.Errorf("invalid signature: %w", err)
// 	}

// 	// Parse public key
// 	pubkey, err := deserializeBLSPublicKey(pubkeyData)
// 	if err != nil {
// 		return false, fmt.Errorf("invalid public key: %w", err)
// 	}

// 	// Hash the message
// 	messageHash := crypto.Keccak256Hash([]byte(message))
// 	var messageArray [32]byte
// 	copy(messageArray[:], messageHash[:])

// 	// Verify using BLS
// 	return sig.Verify(pubkey, messageArray)
// }

// func VerifyBLSSignatureFromJSON(jsonData interface{}, signature string, pubkeyData string) (bool, error) {
// 	jsonDataMap := make(map[string]interface{})

// 	jsonBytes, err := json.Marshal(jsonData)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to marshal input data: %w", err)
// 	}

// 	if err := json.Unmarshal(jsonBytes, &jsonDataMap); err != nil {
// 		return false, fmt.Errorf("failed to unmarshal to map: %w", err)
// 	}

// 	convertToLower(jsonDataMap)

// 	jsonDataBytes, err := json.Marshal(jsonDataMap)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to marshal json data: %w", err)
// 	}

// 	message := string(jsonDataBytes)
// 	return VerifyBLSSignature(message, signature, pubkeyData)
// }

// // Utility functions for key generation and management

// func GenerateBLSKeyPair() (*bls.KeyPair, error) {
// 	return bls.GenRandomBlsKeys()
// }

// func ReadBLSKeyFromFile(keyPath string, password string) (*bls.KeyPair, error) {
// 	return bls.ReadPrivateKeyFromFile(keyPath, password)
// }

// func deserializeBLSSignature(data string) (*bls.Signature, error) {
// 	// If the data is in JSON format, parse it
// 	if strings.HasPrefix(data, "{") && strings.HasSuffix(data, "}") {
// 		var sigJSON struct {
// 			X string `json:"x"`
// 			Y string `json:"y"`
// 		}
// 		if err := json.Unmarshal([]byte(data), &sigJSON); err != nil {
// 			return nil, fmt.Errorf("failed to parse JSON signature: %w", err)
// 		}

// 		// Parse X coordinate
// 		xBI := new(big.Int)
// 		if strings.HasPrefix(sigJSON.X, "0x") {
// 			_, ok := xBI.SetString(sigJSON.X[2:], 16)
// 			if !ok {
// 				return nil, fmt.Errorf("failed to parse X coordinate")
// 			}
// 		} else {
// 			_, ok := xBI.SetString(sigJSON.X, 16)
// 			if !ok {
// 				return nil, fmt.Errorf("failed to parse X coordinate")
// 			}
// 		}

// 		// Parse Y coordinate
// 		yBI := new(big.Int)
// 		if strings.HasPrefix(sigJSON.Y, "0x") {
// 			_, ok := yBI.SetString(sigJSON.Y[2:], 16)
// 			if !ok {
// 				return nil, fmt.Errorf("failed to parse Y coordinate")
// 			}
// 		} else {
// 			_, ok := yBI.SetString(sigJSON.Y, 16)
// 			if !ok {
// 				return nil, fmt.Errorf("failed to parse Y coordinate")
// 			}
// 		}

// 		// Create G1 point and wrap in signature
// 		g1Point := bls.NewG1Point(xBI, yBI)
// 		return &bls.Signature{G1Point: g1Point}, nil
// 	}

// 	// Otherwise treat as hex string for backward compatibility
// 	var signatureBytes []byte
// 	_, err := fmt.Sscanf(data, "%x", &signatureBytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode signature hex: %w", err)
// 	}

// 	// Use BLS library's deserialization
// 	g1Point := bls.NewZeroG1Point()
// 	g1Point = g1Point.Deserialize(signatureBytes)
// 	return &bls.Signature{G1Point: g1Point}, nil
// }

// func serializeBLSPublicKey(pubkey *bls.G2Point) string {
// 	// Use the BLS library's serialization method
// 	serialized := pubkey.Serialize()
// 	return fmt.Sprintf("%x", serialized)
// }

// func deserializeBLSPublicKey(data string) (*bls.G2Point, error) {
// 	// Convert hex string to bytes
// 	var pubkeyBytes []byte
// 	_, err := fmt.Sscanf(data, "%x", &pubkeyBytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode public key hex: %w", err)
// 	}

// 	// Use BLS library's deserialization
// 	pubkey := bls.NewZeroG2Point()
// 	pubkey = pubkey.Deserialize(pubkeyBytes)
// 	return pubkey, nil
// }

// // Extension methods for KeyPair to match ECDSA interface
// func GetBLSPrivateKeyString(keyPair *bls.KeyPair) string {
// 	return keyPair.PrivKey.String()
// }

// func GetBLSPublicKeyString(keyPair *bls.KeyPair) string {
// 	return serializeBLSPublicKey(keyPair.GetPubKeyG2())
// }

// // Convert hex private key to decimal string format required by BLS library
// func HexToBLSPrivateKey(hexKey string) (string, error) {
// 	// Remove "0x" prefix if present
// 	if len(hexKey) > 2 && hexKey[:2] == "0x" {
// 		hexKey = hexKey[2:]
// 	}

// 	// Parse hex to big.Int
// 	bigIntValue := new(big.Int)
// 	_, success := bigIntValue.SetString(hexKey, 16)
// 	if !success {
// 		return "", fmt.Errorf("failed to parse hex key")
// 	}

// 	// Convert to decimal string
// 	return bigIntValue.String(), nil
// }
