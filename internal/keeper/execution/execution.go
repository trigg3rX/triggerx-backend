package execution

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"time"

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
	logger.Info("[Execution] Executing Task")

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

	taskDefinitionId := 0
	if val, ok := requestBody["taskDefinitionId"].(float64); ok {
		taskDefinitionId = int(val)
	}
	logger.Info("[Execution] taskDefinitionId: %v\n", taskDefinitionId)

	jobData, ok := requestBody["job"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job data"})
		return
	}

	// Construct job object from request data with safe defaults
	job := &types.Job{
		JobID:          0,
		TargetFunction: "",
		Arguments:      map[string]interface{}{},
		ChainID:        0,
	}
	if id, ok := jobData["job_id"].(int64); ok {
		job.JobID = id
	}
	if tf, ok := jobData["targetFunction"].(string); ok {
		job.TargetFunction = tf
	}
	if args, ok := jobData["arguments"].(map[string]interface{}); ok {
		job.Arguments = args
	}
	if chain, ok := jobData["chainID"].(int); ok {
		job.ChainID = chain
	}
	if ca, ok := jobData["contractAddress"].(string); ok {
		job.TargetContractAddress = ca
	}

	// Execute job and handle any execution errors
	jobExecutor := NewJobExecutor()
	execResult, err := jobExecutor.Execute(job)
	if err != nil {
		logger.Error("[Execution] Error executing job:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

	execResultBytes, err := json.Marshal(execResult)
	if err != nil {
		logger.Error("[Execution] Error marshaling execution result:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
		return
	}
	krw := &keeperResponseWrapper{Data: execResultBytes}

	// Build proof template with execution metadata and placeholders
	proofTemplate := proof.ProofTemplate{
		JobID:            job.JobID,
		JobType:          job.TargetFunction,
		TaskID:           job.JobID,
		TaskDefinitionID: int64(taskDefinitionId),
		Trigger: proof.TriggerInfo{
			Timestamp:         time.Now().UTC().Format(time.RFC3339),
			Value:             "triggered",
			TxHash:            "0x",
			EventName:         "Event",
			ConditionEndpoint: "http://example.com/condition",
			ConditionValue:    "value",
			CustomTriggerDefinition: proof.CustomTriggerInfo{
				Type:   "custom",
				Params: map[string]interface{}{"example": "param"},
			},
		},
		Action: proof.ActionInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			TxHash:    "0x",
			GasUsed:   "0",
			Status:    "success",
		},
	}

	// Mock TLS state for proof generation
    certBytes := []byte("mock certificate data")
    mockCert := &x509.Certificate{Raw: certBytes}
    connState := &tls.ConnectionState{
        PeerCertificates: []*x509.Certificate{mockCert},
    }

	// Generate and store proof on IPFS, returning content identifier (CID)
	cid, err := proof.GenerateAndStoreProof(krw, connState, proofTemplate)
	if err != nil {
		logger.Error("[Execution] Error generating/storing proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	// Generate TLS proof for response verification
	tlsProof, err := proof.GenerateProof(krw, connState)
	if err != nil {
		logger.Error("[Execution] Error generating TLS proof:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	services.SendTask(cid, cid, taskDefinitionId)

	logger.Info("[Execution] CID: ", cid)

	// Combine all response data for attester
	responseData := map[string]interface{}{
		"executionResult":  execResult,
		"proof":            tlsProof,
		"cid":              cid,
		"taskDefinitionId": taskDefinitionId,
	}

	responseBytes, err := json.Marshal(responseData)
	if err != nil {
		logger.Error("[Execution] Error marshaling response:", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", responseBytes)
}
