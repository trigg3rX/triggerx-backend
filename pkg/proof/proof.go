package proof

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TLSProof struct holds proof details including the BLS signature
type TLSProof struct {
	CertificateHash string `json:"certificateHash"`
	ResponseHash    string `json:"responseHash"`
	Timestamp       string `json:"timestamp"`
}

type PinataConfig struct {
	APIKey    string
	SecretKey string
	Host      string
}

// KeeperResponse interface to handle different response types
type KeeperResponse interface {
	GetData() []byte
}

func LoadPinataConfig() (*PinataConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("PINATA_API_KEY")
	secretKey := os.Getenv("PINATA_SECRET_API_KEY")
	host := os.Getenv("IPFS_HOST")

	if apiKey == "" || secretKey == "" || host == "" {
		return nil, errors.New("missing required Pinata configuration")
	}

	return &PinataConfig{
		APIKey:    apiKey,
		SecretKey: secretKey,
		Host:      host,
	}, nil
}

// GenerateProof creates a TLS proof with a BLS signature
func GenerateProof(response KeeperResponse, connState *tls.ConnectionState) (*TLSProof, error) {
	if connState == nil || len(connState.PeerCertificates) == 0 {
		return nil, errors.New("no TLS certificates found")
	}

	// Hash the first certificate
	certHash := sha256.Sum256(connState.PeerCertificates[0].Raw)
	certHashStr := hex.EncodeToString(certHash[:])

	// Hash the response data
	respHash := sha256.Sum256(response.GetData())
	respHashStr := hex.EncodeToString(respHash[:])

	// Create proof with current timestamp
	return &TLSProof{
		CertificateHash: certHashStr,
		ResponseHash:    respHashStr,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func GenerateAndStoreProof(
	response KeeperResponse,
	connState *tls.ConnectionState,
	tempData types.IPFSData,
) (types.IPFSData, error) {
	// Generate the proof
	proof, err := GenerateProof(response, connState)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to generate proof: %v", err)
	}

	tempData.ProofData.TaskID = tempData.ActionData.TaskID
	tempData.ProofData.Timestamp = time.Now()
	tempData.ProofData.CertificateHash = proof.CertificateHash
	tempData.ProofData.ResponseHash = proof.ResponseHash

	// Convert template to JSON
	jsonData, err := json.MarshalIndent(tempData, "", "  ")
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to marshal template: %v", err)
	}

	// Load Pinata config
	pinataConfig, err := LoadPinataConfig()
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to load Pinata config: %v", err)
	}

	// Create unique name for the file
	fileName := fmt.Sprintf("proof_%d_%d_%s.json",
		tempData.JobData.JobID,
		tempData.ActionData.TaskID,
		time.Now().UTC().Format(time.RFC3339))

	// Store in IPFS using Pinata with metadata
	ipfsData, err := uploadToPinata(tempData, jsonData, fileName, pinataConfig)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to upload to Pinata: %v", err)
	}

	return ipfsData, nil
}

// uploadToPinata uploads the data to IPFS using Pinata API
func uploadToPinata(tempData types.IPFSData, data []byte, fileName string, config *PinataConfig) (types.IPFSData, error) {
	url := "https://api.pinata.cloud/pinning/pinJSONToIPFS"

	// Create metadata for the file
	metadata := map[string]interface{}{
		"pinataMetadata": map[string]interface{}{
			"name": fileName,
		},
		"pinataContent": json.RawMessage(data),
	}

	// Convert metadata to JSON
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to marshal metadata: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("pinata_api_key", config.APIKey)
	req.Header.Set("pinata_secret_api_key", config.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.IPFSData{}, fmt.Errorf("failed to upload to Pinata: status %d", resp.StatusCode)
	}

	var result struct {
		IpfsHash string `json:"IpfsHash"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to decode response: %v", err)
	}

	var ipfsData types.IPFSData = tempData
	ipfsData.ProofData.ActionDataCID = result.IpfsHash

	return ipfsData, nil
}
