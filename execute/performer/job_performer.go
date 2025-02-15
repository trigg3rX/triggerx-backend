package performer

import (
	// "crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	// Import your trigger project packages (adjust the module paths as needed)
	"github.com/trigg3rX/triggerx-backend/execute/keeper/executor"
	"github.com/trigg3rX/triggerx-backend/execute/manager"
	"github.com/trigg3rX/triggerx-backend/pkg/proof"
)

// keeperResponseWrapper implements the KeeperResponse interface from the proof module.
type keeperResponseWrapper struct {
	Data []byte
}

// GetData returns the underlying data bytes.
func (krw *keeperResponseWrapper) GetData() []byte {
	return krw.Data
}

// ExecuteTask handles incoming task requests, executes the job, generates a proof,
// and sends the proof response (as bytes) back to the attester.
func ExecuteTask(c *gin.Context) {
	log.Println("Executing Task")

	// Only allow POST requests.
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Invalid method",
		})
		return
	}

	// Parse the JSON request body.
	var requestBody map[string]interface{}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}

	// Extract taskDefinitionId if provided.
	taskDefinitionId := 0
	if val, ok := requestBody["taskDefinitionId"].(float64); ok {
		taskDefinitionId = int(val)
	}
	log.Printf("taskDefinitionId: %v\n", taskDefinitionId)

	// Expect job details to be provided under the "job" key.
	jobData, ok := requestBody["job"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing job data"})
		return
	}

	// Create a job object from the job data.
	job := &manager.Job{
		JobID:          "",
		TargetFunction: "",
		Arguments:      map[string]interface{}{},
		ChainID:        "",
	}
	if id, ok := jobData["job_id"].(string); ok {
		job.JobID = id
	}
	if tf, ok := jobData["targetFunction"].(string); ok {
		job.TargetFunction = tf
	}
	if args, ok := jobData["arguments"].(map[string]interface{}); ok {
		job.Arguments = args
	}
	if chain, ok := jobData["chainID"].(string); ok {
		job.ChainID = chain
	}
	if ca, ok := jobData["contractAddress"].(string); ok {
		job.ContractAddress = ca
	}

	// Execute the job using your custom JobExecutor.
	jobExecutor := executor.NewJobExecutor()
	execResult, err := jobExecutor.Execute(job)
	if err != nil {
		log.Println("Error executing job:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Job execution failed"})
		return
	}

	// Marshal the execution result into JSON bytes.
	execResultBytes, err := json.Marshal(execResult)
	if err != nil {
		log.Println("Error marshaling execution result:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal execution result"})
		return
	}
	krw := &keeperResponseWrapper{Data: execResultBytes}

	// Prepare the proof template with job and trigger/action details.
	proofTemplate := proof.ProofTemplate{
		JobID:            job.JobID,
		JobType:          job.TargetFunction,
		TaskID:           fmt.Sprintf("task-%s", job.JobID),
		TaskDefinitionID: fmt.Sprintf("%d", taskDefinitionId),
		Trigger: proof.TriggerInfo{
			Timestamp:         time.Now().UTC().Format(time.RFC3339),
			Value:             "triggered",
			TxHash:            "0x", // placeholder; replace as needed
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
			TxHash:    "0x", // placeholder; replace as needed
			GasUsed:   "0", // placeholder; replace as needed
			Status:    "success",
		},
	}

	// Get the TLS connection state from the request.
	connState := c.Request.TLS

	// Generate and store the proof.
	// This will return a CID (e.g. from Pinata) which is our stored proof identifier.
	cid, err := proof.GenerateAndStoreProof(krw, connState, proofTemplate)
	if err != nil {
		log.Println("Error generating/storing proof:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	// Also generate the TLS proof directly to obtain the proof (hash) details.
	tlsProof, err := proof.GenerateProof(krw, connState)
	if err != nil {
		log.Println("Error generating TLS proof:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proof generation failed"})
		return
	}

	// Prepare the final response including the execution result, proof details, and CID.
	responseData := map[string]interface{}{
		"executionResult":  execResult,
		"proof":            tlsProof, // includes certificateHash, responseHash, and timestamp
		"cid":              cid,
		"taskDefinitionId": taskDefinitionId,
	}

	// Marshal the response data into JSON bytes.
	responseBytes, err := json.Marshal(responseData)
	if err != nil {
		log.Println("Error marshaling response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	// Send the response as raw bytes to the attester.
	c.Data(http.StatusOK, "application/octet-stream", responseBytes)
}
