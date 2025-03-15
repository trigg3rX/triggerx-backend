package manager

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/internal/manager/scheduler"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

var (
	logger       logging.Logger
	jobScheduler *scheduler.JobScheduler

	EtherscanApiKey string
	AlchemyApiKey   string

	DeployerPrivateKey string

	ManagerRPCPort    string
	ManagerIPAddress  string
	DatabaseIPAddress string
)

func Init() {
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)

	// var err error
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	jobScheduler, err = scheduler.NewJobScheduler(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize job scheduler: %v", err)
	}

	EtherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	AlchemyApiKey = os.Getenv("ALCHEMY_API_KEY")
	DeployerPrivateKey = os.Getenv("PRIVATE_KEY_DEPLOYER")
	ManagerRPCPort = os.Getenv("MANAGER_RPC_PORT")
	ManagerIPAddress = os.Getenv("MANAGER_IP_ADDRESS")
	DatabaseIPAddress = os.Getenv("DATABASE_IP_ADDRESS")

	if EtherscanApiKey == "" || AlchemyApiKey == "" || DeployerPrivateKey == "" || ManagerIPAddress == "" || ManagerRPCPort == "" || DatabaseIPAddress == "" {
		logger.Fatal(".env VARIABLES NOT SET PROPERLY !!!")
	}

	gin.SetMode(gin.ReleaseMode)
}

func HandleCreateJobEvent(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Invalid method"})
		return
	}

	var jobData types.HandleCreateJobData
	if err := c.BindJSON(&jobData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	var err error
	switch jobData.TaskDefinitionID {
	case 1, 2: // Time-based jobs
		err = jobScheduler.StartTimeBasedJob(jobData)
	case 3, 4: // Event-based jobs
		err = jobScheduler.StartEventBasedJob(jobData)
	case 5, 6: // Condition-based jobs
		err = jobScheduler.StartConditionBasedJob(jobData)
	default:
		logger.Warnf("Unknown task definition ID: %d for job: %d",
			jobData.TaskDefinitionID, jobData.JobID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition ID"})
		return
	}

	if err != nil {
		logger.Errorf("Failed to schedule job %d: %v", jobData.JobID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule job"})
		return
	}

	logger.Infof("Successfully scheduled job with ID: %d", jobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job scheduled successfully"})
}

func HandleUpdateJobEvent(c *gin.Context) {
	var updateJobData types.HandleUpdateJobData
	if err := c.BindJSON(&updateJobData); err != nil {
		logger.Error("Failed to parse update job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// TODO: Implement job update logic using scheduler
	logger.Infof("Job update requested for ID: %d", updateJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job update request received"})
}

func HandlePauseJobEvent(c *gin.Context) {
	var pauseJobData types.HandlePauseJobData
	if err := c.BindJSON(&pauseJobData); err != nil {
		logger.Error("Failed to parse pause job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// TODO: Implement job pause logic using scheduler
	logger.Infof("Job pause requested for ID: %d", pauseJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job pause request received"})
}

func HandleResumeJobEvent(c *gin.Context) {
	var resumeJobData types.HandleResumeJobData
	if err := c.BindJSON(&resumeJobData); err != nil {
		logger.Error("Failed to parse resume job data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// TODO: Implement job resume logic using scheduler
	logger.Infof("Job resume requested for ID: %d", resumeJobData.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Job resume request received"})
}

func HandleKeeperConnectEvent(c *gin.Context) {
	var keeperData types.UpdateKeeperConnectionData
	if err := c.BindJSON(&keeperData); err != nil {
		logger.Error("Failed to parse keeper data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	url := "http://localhost:8080/api/keepers/connection"

	jsonData, err := json.Marshal(keeperData)
	if err != nil {
		logger.Error("Failed to marshal keeper data", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal keeper data"})
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create HTTP request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create HTTP request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	var response types.UpdateKeeperConnectionDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error("Failed to decode response", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode response"})	
		return
	}

	logger.Infof("Keeper connected: %s", keeperData.KeeperAddress)
	c.JSON(http.StatusOK, response)
}
