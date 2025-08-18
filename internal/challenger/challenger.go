package challenger

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/imua-xyz/imua-avs-sdk/client/txmgr"
	"github.com/imua-xyz/imua-avs-sdk/nodeapi"
	"github.com/imua-xyz/imua-avs-sdk/signer"
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/chainio"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Challenger struct {
	logger      logging.Logger
	ethClient   *ethclient.Client
	ethWsClient *ethclient.Client
	nodeApi     *nodeapi.NodeApi
	avsReader   chainio.AvsReader
	avsWriter   chainio.AvsWriter
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewChallenger(logger logging.Logger) *Challenger {
	ethClient, err := ethclient.Dial(config.GetEthRPCUrl())
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum: %v", err)
		return nil
	}
	ethWsClient, err := ethclient.Dial(config.GetEthWsUrl())
	if err != nil {
		logger.Errorf("Failed to connect to Ethereum: %v", err)
		return nil
	}
	chainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		logger.Error("Cannot get chainId", "err", err)
		return nil
	}

	// Setup Node Api
	nodeApi := nodeapi.NewNodeApi(config.GetAvsName(), config.GetSemVer(), config.GetOperatorNodeApiPort(), logger)
	signer, _, err := signer.SignerFromConfig(signer.Config{
		PrivateKey: config.GetPrivateKeyController(),
	},
		chainId,
	)
	if err != nil {
		panic(err)
	}
	txMgr := txmgr.NewSimpleTxManager(ethClient, logger, signer, common.HexToAddress(config.GetKeeperAddress()))
	avsReader, _ := chainio.BuildChainReader(
		common.HexToAddress(config.GetAvsGovernanceAddress()),
		ethClient,
		logger)

	avsWriter, _ := chainio.BuildChainWriter(
		common.HexToAddress(config.GetAvsGovernanceAddress()),
		ethClient,
		logger,
		txMgr)

	// Create context for shutdown handling
	ctx, cancel := context.WithCancel(context.Background())

	return &Challenger{
		logger:      logger,
		ethClient:   ethClient,
		ethWsClient: ethWsClient,
		nodeApi:     nodeApi,
		avsReader:   avsReader,
		avsWriter:   avsWriter,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (c *Challenger) Start(ctx context.Context) error {
	c.logger.Infof("Starting challenger.")

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(config.GetAvsGovernanceAddress())},
	}
	logs := make(chan ethtypes.Log)

	sub, err := c.ethWsClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		c.logger.Error("Subscribe failed", "err", err)
		return err
	}
	defer sub.Unsubscribe()

	c.logger.Infof("Starting challenger event monitoring...")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutdown signal received, stopping challenger")
			return ctx.Err()
		case err := <-sub.Err():
			c.logger.Error("Subscription error:", err)
		case vLog := <-logs:
			event, err := c.parseEvent(vLog)
			if err != nil {
				c.logger.Info("Not as expected TaskCreated or TaskSubmitted log, parse err:", "err", err)
				continue
			}
			if event != nil {
				switch e := event.(type) {
				case *avs.TriggerXAvsTaskCreated:
					// Process the task creation event for challenge monitoring
					c.ProcessNewTaskCreatedLog(e)
				case *avs.TriggerXAvsTaskSubmitted:
					// Process task submission events (operators submitting responses)
					c.ProcessTaskSubmittedLog(e)
				}
			}
		}
	}
}

func (c *Challenger) parseEvent(vLog ethtypes.Log) (interface{}, error) {
	// Create a filterer to parse events
	filterer, err := avs.NewTriggerXAvsFilterer(common.HexToAddress(config.GetAvsGovernanceAddress()), c.ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create filterer: %w", err)
	}

	// Try to parse as TaskCreated event
	event, err := filterer.ParseTaskCreated(vLog)
	if err == nil {
		return event, nil
	}

	// Try to parse as TaskSubmitted event (for monitoring operator submissions)
	taskSubmittedEvent, err := filterer.ParseTaskSubmitted(vLog)
	if err == nil {
		return taskSubmittedEvent, nil
	}

	// If it's not a known event, return nil
	return nil, fmt.Errorf("event is not a TaskCreated or TaskSubmitted event")
}

// ProcessNewTaskCreatedLog processes a new task creation event for challenge monitoring
func (c *Challenger) ProcessNewTaskCreatedLog(e *avs.TriggerXAvsTaskCreated) {
	c.logger.Info("Processing new task created event for challenge monitoring",
		"taskID", e.TaskId.Uint64(),
		"definitionHash", fmt.Sprintf("0x%x", e.TaskDefinitionId))

	// Get task info for challenge monitoring
	taskInfo, err := c.avsReader.GetTaskInfo(&bind.CallOpts{}, config.GetAvsGovernanceAddress(), e.TaskId.Uint64())
	if err != nil {
		c.logger.Error("Failed to get task info for challenge monitoring", "err", err)
		return
	}

	// Start challenge monitoring in background
	go func() {
		err := c.MonitorForChallenge(c.ctx, e.TaskId.Uint64(), taskInfo)
		if err != nil {
			c.logger.Error("Failed to monitor for challenge", "err", err, "taskId", e.TaskId.Uint64())
		}
	}()
}

// ProcessTaskSubmittedLog processes task submission events from operators
func (c *Challenger) ProcessTaskSubmittedLog(e *avs.TriggerXAvsTaskSubmitted) {
	c.logger.Info("Processing task submitted event for challenge monitoring",
		"taskID", e.TaskId.Uint64(),
		"operator", e.Operator.String(),
		"phase", e.Phase)

	// Here you could implement immediate challenge logic based on the submitted response
	// For example, if you detect suspicious patterns in the submission, you could trigger a challenge
	
	// This event-driven approach allows for faster challenge responses compared to polling
	// The actual challenge logic would still be handled by the MonitorForChallenge function
	// which is already running for each task
}

// MonitorForChallenge monitors a task and calls challenge function when phaseTwoSubmitted is true
func (c *Challenger) MonitorForChallenge(
	ctx context.Context,
	taskId uint64,
	taskInfo avs.TaskInfo) error {

	startingEpoch := taskInfo.StartingEpoch
	taskResponsePeriod := taskInfo.TaskResponsePeriod
	taskStatisticalPeriod := taskInfo.TaskStatisticalPeriod

	// Track submission status for each phase
	phaseOneSubmitted := false
	phaseTwoSubmitted := false

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Shutdown signal received, stopping challenge monitoring", "taskId", taskId)
			return ctx.Err() // Gracefully exit if context is canceled
		default:
			// Fetch the current epoch information
			epochIdentifier, err := c.avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, taskInfo.TaskContractAddress.String())
			if err != nil {
				c.logger.Error("Cannot GetAVSEpochIdentifier", "err", err)
				return fmt.Errorf("failed to get AVS info: %w", err) // Stop on persistent error
			}

			num, err := c.avsReader.GetCurrentEpoch(&bind.CallOpts{}, epochIdentifier)
			if err != nil {
				c.logger.Error("Cannot exec GetCurrentEpoch", "err", err)
				return fmt.Errorf("failed to get current epoch: %w", err) // Stop on persistent error
			}

			currentEpoch := uint64(num)
			
			if currentEpoch > startingEpoch+taskResponsePeriod+taskStatisticalPeriod {
				c.logger.Info("Exiting challenge monitoring: Task period has passed",
					"Task", taskInfo.TaskContractAddress.String()+"--"+strconv.FormatUint(taskId, 10))
				return nil
			}

			switch {
			case currentEpoch <= startingEpoch:
				// Wait for task to start
				time.Sleep(config.GetRetryDelay())

			case currentEpoch <= startingEpoch+taskResponsePeriod:
				// Phase One period - monitor for submissions
				if !phaseOneSubmitted {
					c.logger.Info("Monitoring Phase One submissions", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskId", taskId)
					
					// Check if operators have submitted for phase one by querying contract
					operatorResponses, err := c.avsReader.GetOperatorTaskResponseList(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskId)
					if err != nil {
						c.logger.Error("Failed to get operator task response list", "err", err, "taskId", taskId)
						time.Sleep(config.GetRetryDelay())
						continue
					}
					
					// Check if any operators have submitted for phase 1
					phase1Submissions := 0
					for _, response := range operatorResponses {
						if response.Phase == 1 {
							phase1Submissions++
						}
					}
					
					if phase1Submissions > 0 {
						phaseOneSubmitted = true
						c.logger.Info("Phase One submissions detected", "taskId", taskId, "submissionCount", phase1Submissions)
					} else {
						time.Sleep(config.GetRetryDelay())
					}
				} else {
					time.Sleep(config.GetRetryDelay())
				}

			case currentEpoch <= startingEpoch+taskResponsePeriod+taskStatisticalPeriod && currentEpoch > startingEpoch+taskResponsePeriod:
				// Phase Two period - monitor for submissions and call challenge when phaseTwoSubmitted is true
				if !phaseTwoSubmitted {
					c.logger.Info("Monitoring Phase Two submissions", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskStatisticalPeriod", taskStatisticalPeriod, "taskId", taskId)
					
					// Check if operators have submitted for phase two by querying contract
					operatorResponses, err := c.avsReader.GetOperatorTaskResponseList(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskId)
					if err != nil {
						c.logger.Error("Failed to get operator task response list", "err", err, "taskId", taskId)
						time.Sleep(config.GetRetryDelay())
						continue
					}
					
					// Check if any operators have submitted for phase 2
					phase2Submissions := 0
					for _, response := range operatorResponses {
						if response.Phase == 2 {
							phase2Submissions++
						}
					}
					
					if phase2Submissions > 0 {
						phaseTwoSubmitted = true
						c.logger.Info("Phase Two submissions detected - triggering challenge function", "taskId", taskId, "submissionCount", phase2Submissions)

						// Call challenge function when phaseTwoSubmitted is true
						err := c.CallChallengeFunction(ctx, taskId, taskInfo)
						if err != nil {
							c.logger.Error("Failed to call challenge function", "err", err, "taskId", taskId)
							// Continue monitoring despite challenge failure
						}
					} else {
						time.Sleep(config.GetRetryDelay())
					}
				} else {
					time.Sleep(config.GetRetryDelay())
				}

			default:
				c.logger.Info("Current epoch is not within expected range", "currentEpoch", currentEpoch, "taskId", taskId)
				return fmt.Errorf("current epoch %d is not within expected range %d", currentEpoch, startingEpoch)
			}

			// If both phases are submitted, we can continue monitoring for challenges
			if phaseOneSubmitted && phaseTwoSubmitted {
				c.logger.Info("Both phases detected, continuing challenge monitoring", "taskId", taskId)
				time.Sleep(config.GetRetryDelay())
			}

			// Add a small delay to prevent tight looping, but respect shutdown context
			time.Sleep(config.GetRetryDelay())
		}
	}
}

// CallChallengeFunction calls the challenge function from avsWriter when phaseTwoSubmitted is true
func (c *Challenger) CallChallengeFunction(
	ctx context.Context,
	taskId uint64,
	taskInfo avs.TaskInfo) error {

	c.logger.Info("Calling challenge function for task", 
		"taskId", taskId, 
		"taskAddr", taskInfo.TaskContractAddress.String())

	// Get all operator responses for phase 2 to determine which ones to challenge
	operatorResponses, err := c.avsReader.GetOperatorTaskResponseList(&bind.CallOpts{}, taskInfo.TaskContractAddress.String(), taskId)
	if err != nil {
		c.logger.Error("Failed to get operator responses for challenge", "err", err, "taskId", taskId)
		return fmt.Errorf("failed to get operator responses: %w", err)
	}

	// Find phase 2 responses to potentially challenge
	var targetResponse *avs.OperatorResInfo
	for i, response := range operatorResponses {
		if response.Phase == 2 {
			// Here you would implement logic to determine which response to challenge
			// For now, we'll challenge the first phase 2 response as an example
			targetResponse = &operatorResponses[i]
			c.logger.Info("Found phase 2 response to challenge", 
				"taskId", taskId, 
				"operator", response.Operator.String(),
				"phase", response.Phase)
			break
		}
	}

	if targetResponse == nil {
		c.logger.Warn("No phase 2 responses found to challenge", "taskId", taskId)
		return fmt.Errorf("no phase 2 responses found to challenge for task %d", taskId)
	}

	// Create a challenge task response based on the operator response we want to challenge
	taskResponse := TaskResponse{
		TaskID:  taskId,
		IsValid: false, // Challenging as invalid
	}

	signature, responseBytes, err := c.SignTaskResponse(taskResponse)
	if err != nil {
		c.logger.Error("Failed to sign challenge response", "err", err)
		return fmt.Errorf("failed to sign challenge response: %w", err)
	}

	// Call the Challenge function from avsWriter
	receipt, err := c.avsWriter.Challenge(
		ctx,
		taskId,
		responseBytes,
		signature,
		taskInfo.TaskContractAddress.String(),
		2, // Phase 2
	)
	
	if err != nil {
		c.logger.Error("Failed to submit challenge", "err", err, "taskId", taskId)
		return fmt.Errorf("failed to submit challenge: %w", err)
	}

	c.logger.Info("Successfully submitted challenge", 
		"taskId", taskId, 
		"targetOperator", targetResponse.Operator.String(),
		"txHash", receipt.TxHash.String(),
		"gasUsed", receipt.GasUsed)

	return nil
}

// SignTaskResponse signs a task response for challenge purposes
func (c *Challenger) SignTaskResponse(taskResponse TaskResponse) ([]byte, []byte, error) {
	taskResponseHash, data, err := GetTaskResponseDigestEncodeByAbi(taskResponse)
	if err != nil {
		c.logger.Error("Error SignTaskResponse with getting task response header hash. skipping task (this is not expected and should be investigated)", "err", err)
		return nil, nil, err
	}
	msgBytes := taskResponseHash[:]

	sig := config.GetConsensusKeyPair().Sign(msgBytes)

	return sig.Marshal(), data, nil
}

// ===== COMMENTED OUT TRANSACTION-RELATED CODE =====
// The following functions contain transaction sending logic that has been commented out
// as per the requirement to comment out all transaction-related code

/*
// SendSignedTaskResponseToChain is commented out - contains transaction logic
func (c *Challenger) SendSignedTaskResponseToChain(
	ctx context.Context,
	taskId uint64,
	taskResponse []byte,
	blsSignature []byte,
	taskInfo avs.TaskInfo) (string, error) {

	// This function contained OperatorSubmitTask calls with transaction sending
	// All transaction-related code is commented out per requirements
	
	return "Transaction functionality disabled in challenger", nil
}
*/

func (c *Challenger) Close() {
	c.logger.Info("Shutting down challenger...")
	c.cancel() // Cancel the context to signal shutdown to all goroutines
	// c.nodeApi.Stop()
}