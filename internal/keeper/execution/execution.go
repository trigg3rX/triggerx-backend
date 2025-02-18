package execution

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
	"github.com/trigg3rX/triggerx-backend/pkg/types"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/services"
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

	jobData := requestBody["job"].(types.Job)
	triggerData := requestBody["trigger"].(types.TriggerData)

	logger.Infof("taskDefinitionId: %v\n", jobData.TaskDefinitionID)

	// Execute job and handle any execution errors
	jobExecutor := NewJobExecutor()
	actionData, err := jobExecutor.Execute(&jobData)
	if err != nil {
		logger.Errorf("Error executing job:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

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
		JobData: jobData,
		TriggerData: triggerData,
		ActionData: actionData,
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

	services.SendTask(tlsProof.ResponseHash, ipfsData.ProofData.ActionDataCID, jobData.TaskDefinitionID)

	logger.Infof("CID: %s", ipfsData.ProofData.ActionDataCID)

	ipfsDataBytes, err := json.Marshal(ipfsData)
	if err != nil {
		logger.Errorf("Error marshaling IPFS data:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal IPFS data"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", ipfsDataBytes)
}
