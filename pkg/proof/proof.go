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
)

type ProofTemplate struct {
	JobID            string      `json:"job_id"`
	JobType          string      `json:"job_type"`
	TaskID           string      `json:"task_id"`
	TaskDefinitionID string      `json:"task_definition_id"`
	Trigger          TriggerInfo `json:"trigger"`
	Action           ActionInfo  `json:"action"`
	Proof            *TLSProof   `json:"proof"`
}

type TriggerInfo struct {
	Timestamp               string            `json:"timestamp"`
	Value                   string            `json:"value"`
	TxHash                  string            `json:"txHash"`
	EventName               string            `json:"eventName"`
	ConditionEndpoint       string            `json:"conditionEndpoint"`
	ConditionValue          string            `json:"conditionValue"`
	CustomTriggerDefinition CustomTriggerInfo `json:"customTriggerDefinition"`
}

type CustomTriggerInfo struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

type ActionInfo struct {
	Timestamp string `json:"timestamp"`
	TxHash    string `json:"txHash"`
	GasUsed   string `json:"gasUsed"`
	Status    string `json:"status"`
}

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
	templateData ProofTemplate,
) (string, error) {
	// Generate the proof
	proof, err := GenerateProof(response, connState)
	if err != nil {
		return "", fmt.Errorf("failed to generate proof: %v", err)
	}

	// Add proof to template
	templateData.Proof = proof

	// Set current timestamp if not provided
	if templateData.Action.Timestamp == "" {
		templateData.Action.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Convert template to JSON
	jsonData, err := json.MarshalIndent(templateData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal template: %v", err)
	}

	// Load Pinata config
	pinataConfig, err := LoadPinataConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load Pinata config: %v", err)
	}

	// Create unique name for the file
	fileName := fmt.Sprintf("proof_%s_%s_%s.json",
		templateData.JobID,
		templateData.TaskID,
		time.Now().UTC().Format("20060102150405"))

	// Store in IPFS using Pinata with metadata
	return uploadToPinata(jsonData, fileName, pinataConfig)
}

// uploadToPinata uploads the data to IPFS using Pinata API
func uploadToPinata(data []byte, fileName string, config *PinataConfig) (string, error) {
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
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("pinata_api_key", config.APIKey)
	req.Header.Set("pinata_secret_api_key", config.SecretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload to Pinata: status %d", resp.StatusCode)
	}

	var result struct {
		IpfsHash string `json:"IpfsHash"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.IpfsHash, nil
}
