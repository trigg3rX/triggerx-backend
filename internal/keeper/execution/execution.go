package execution

import (
	// "bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/hex"

	// "fmt"

	"net/http"

	// "os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/aggregator"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

// keeperResponseWrapper wraps execution result bytes to satisfy the proof module's interface
type keeperResponseWrapper struct {
	Data []byte
}

func (krw *keeperResponseWrapper) GetData() []byte {
	return krw.Data
}

func TestAPI(c *gin.Context) {
	logger.Info("Hello Mic testing 1 2 3 ................")

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	var requestBody struct {
		Data string `json:"data"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}
	
	// Handle hex-encoded data (remove "0x" prefix if present)
	hexData := requestBody.Data
	if len(hexData) > 2 && hexData[:2] == "0x" {
		hexData = hexData[2:]
	}
	
	// Decode the hex string to bytes
	decodedData, err := hex.DecodeString(hexData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hex data",
		})
		return
	}
	
	// Convert bytes to string or unmarshal to appropriate structure if needed
	dataString := string(decodedData)

	logger.Infof("requestBody: %v\n", dataString)

	var resp = "Hello"
	c.Data(http.StatusOK, "application/octet-stream", []byte(resp))
}

// ExecuteTask is the main handler for executing keeper tasks. It:
// 1. Validates and processes the incoming job request
// 2. Executes the job and generates execution proof
// 3. Stores proof on IPFS via Pinata
// 4. Returns execution results with proof details to the attester
func ExecuteTask(c *gin.Context) {
	logger.Info("Executing Task")

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	var requestBody map[string]interface{}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}

	// Fix type assertions by properly converting the data
	jobDataRaw, ok := requestBody["job"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job data"})
		return
	}

	triggerDataRaw, ok := requestBody["trigger"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing trigger data"})
		return
	}

	// Convert to proper types
	var jobData types.HandleCreateJobData
	jobDataBytes, err := json.Marshal(jobDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job data format"})
		return
	}
	if err := json.Unmarshal(jobDataBytes, &jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse job data"})
		return
	}
	logger.Infof("jobData: %v\n", jobData)
	var triggerData types.TriggerData
	triggerDataBytes, err := json.Marshal(triggerDataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trigger data format"})
		return
	}
	if err := json.Unmarshal(triggerDataBytes, &triggerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse trigger data"})
		return
	}

	logger.Infof("taskDefinitionId: %v\n", jobData.TaskDefinitionID)

	// Create ethClient using config
	ethClient, err := ethclient.Dial("https://opt-sepolia.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
		return
	}
	defer ethClient.Close()

	// Create job executor with ethClient and etherscan API key
	jobExecutor := NewJobExecutor(ethClient, config.AlchemyAPIKey)

	actionData, err := jobExecutor.Execute(&jobData)
	if err != nil {
		logger.Errorf("Error executing job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

	// // Update keeper metrics after successful job execution
	// keeperID := os.Getenv("KEEPER_ID")
	// if keeperID == "" {
	// 	logger.Warn("KEEPER_ID environment variable not set, using default value")
	// }
	// taskID := triggerData.TaskID

	// // Call the metrics server to store keeper execution metrics
	// if err := StoreKeeperMetrics(keeperID, fmt.Sprintf("%d", taskID)); err != nil {
	// 	logger.Warnf("Failed to store keeper metrics: %v", err)
	// 	// Continue execution even if metrics storage fails
	// } else {
	// 	logger.Infof("Successfully stored metrics for keeper %d and task %d", keeperID, taskID)
	// }

	actionData.TaskID = triggerData.TaskID

	logger.Infof("actionData: %v\n", actionData)

	actionDataBytes, err := json.Marshal(actionData)
	if err != nil {
		logger.Errorf("Error marshaling execution result:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
		return
	}
	krw := &keeperResponseWrapper{Data: actionDataBytes}

	// Mock TLS state for proof generation
	certBytes := []byte("mock certificate data")
	mockCert := &x509.Certificate{Raw: certBytes}
	connState := &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{mockCert},
	}

	tempData := types.IPFSData{
		JobData:     jobData,
		TriggerData: triggerData,
		ActionData:  actionData,
	}

	// Generate and store proof on IPFS, returning content identifier (CID)
	ipfsData, err := proof.GenerateAndStoreProof(krw, connState, tempData)
	if err != nil {
		logger.Errorf("Error generating/storing proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	// Generate TLS proof for response verification
	tlsProof, err := proof.GenerateProof(krw, connState)
	if err != nil {
		logger.Errorf("Error generating TLS proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	aggregator.SendTask(tlsProof.ResponseHash, ipfsData.ProofData.ActionDataCID, jobData.TaskDefinitionID)

	logger.Infof("CID: %s", ipfsData.ProofData.ActionDataCID)

	ipfsDataBytes, err := json.Marshal(ipfsData)
	if err != nil {
		logger.Errorf("Error marshaling IPFS data:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal IPFS data"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", ipfsDataBytes)
}

// StoreKeeperMetrics makes API calls to update keeper metrics in the database
// func StoreKeeperMetrics(keeperID string, taskID string) error {
// 	// Call the increment-tasks endpoint
// 	incrementTasksURL := fmt.Sprintf("http://localhost:8080/api/keepers/%s/increment-tasks", keeperID)
// 	incrementResp, err := http.Post(incrementTasksURL, "application/json", nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to increment keeper task count: %w", err)
// 	}
// 	defer incrementResp.Body.Close()

// 	if incrementResp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("increment task count API returned non-OK status: %d", incrementResp.StatusCode)
// 	}

// 	// Call the add-points endpoint with the task ID
// 	addPointsURL := fmt.Sprintf("http://localhost:8080/api/keepers/%s/add-points", keeperID)

// 	// Create the request payload with the task ID
// 	payload := struct {
// 		TaskID string `
