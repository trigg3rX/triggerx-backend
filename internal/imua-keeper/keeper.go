package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/api/handlers"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/chainio"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Keeper struct {
	logger      logging.Logger
	ethClient   *ethclient.Client
	ethWsClient *ethclient.Client
	nodeApi     *nodeapi.NodeApi
	avsReader   chainio.AvsReader
	avsWriter   chainio.AvsWriter
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewKeeper(logger logging.Logger) *Keeper {
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

	return &Keeper{
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

func (k *Keeper) Start(ctx context.Context) error {
	k.logger.Infof("Starting operator.")
	// k.nodeApi.Start()

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(config.GetAvsGovernanceAddress())},
	}
	logs := make(chan ethtypes.Log)

	sub, err := k.ethWsClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		k.logger.Error("Subscribe failed", "err", err)
		return err
	}
	defer sub.Unsubscribe()

	k.logger.Infof("Starting event monitoring...")

	for {
		select {
		case <-ctx.Done():
			k.logger.Info("Shutdown signal received, stopping keeper")
			return ctx.Err()
		case err := <-sub.Err():
			k.logger.Error("Subscription error:", err)
		case vLog := <-logs:
			event, err := k.parseEvent(vLog)
			if err != nil {
				k.logger.Info("Not as expected TaskCreated log, parse err:", "err", err)
				continue
			}
			if event != nil {
				e := event.(*avs.TriggerXAvsTaskCreated)
				// Process the task creation event
				k.ProcessNewTaskCreatedLog(e)
			}
		}
	}
}

func (k *Keeper) parseEvent(vLog ethtypes.Log) (interface{}, error) {
	// Create a filterer to parse events
	filterer, err := avs.NewTriggerXAvsFilterer(common.HexToAddress(config.GetAvsGovernanceAddress()), k.ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create filterer: %w", err)
	}

	// Try to parse as TaskCreated event
	event, err := filterer.ParseTaskCreated(vLog)
	if err == nil {
		return event, nil
	}

	// If it's not a TaskCreated event, return nil
	return nil, fmt.Errorf("event is not a TaskCreated event")
}

// ProcessNewTaskCreatedLog processes a new task creation event
func (k *Keeper) ProcessNewTaskCreatedLog(e *avs.TriggerXAvsTaskCreated) {
	k.logger.Info("Processing new task created event",
		"taskID", e.TaskId.Uint64(),
		"definitionHash", fmt.Sprintf("0x%x", e.TaskDefinitionId))

	validateRequest := handlers.TaskValidationRequest{
		Data: string(e.TaskData),
	}
	jsonData, err := json.Marshal(validateRequest)
	if err != nil {
		k.logger.Error("Failed to marshal validate request", "err", err)
		return
	}
	request, err := http.NewRequest("POST", "http://localhost:"+config.GetOperatorRPCPort()+"/task/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		k.logger.Error("Failed to create request", "err", err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		k.logger.Error("Failed to send request", "err", err)
		return
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			k.logger.Error("Failed to close response body", "err", err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		k.logger.Error("Failed to read response body", "err", err)
		return
	}

	var validationResponse handlers.ValidationResponse
	err = json.Unmarshal(body, &validationResponse)
	if err != nil {
		k.logger.Error("Failed to unmarshal response body", "err", err)
		return
	}

	if validationResponse.Data {
		k.logger.Info("Task is valid", "taskNumber", e.TaskId.Uint64())
	} else {
		k.logger.Error("Task is invalid", "taskNumber", e.TaskId.Uint64())
	}

	signature, responseBytes, err := k.SignTaskResponse(TaskResponse{
		TaskID:  e.TaskId.Uint64(),
		IsValid: validationResponse.Data,
	})
	if err != nil {
		k.logger.Error("Failed to sign task response", "err", err)
		return
	}

	taskInfo, _ := k.avsReader.GetTaskInfo(&bind.CallOpts{}, config.GetAvsGovernanceAddress(), e.TaskId.Uint64())
	go func() {
		_, err := k.SendSignedTaskResponseToChain(k.ctx, e.TaskId.Uint64(), responseBytes, signature, taskInfo)
		if err != nil {
			k.logger.Error("Failed to send signed task response to chain", "err", err)
		}
	}()
}

func (k *Keeper) SignTaskResponse(taskResponse TaskResponse) ([]byte, []byte, error) {
	taskResponseHash, data, err := GetTaskResponseDigestEncodeByAbi(taskResponse)
	if err != nil {
		k.logger.Error("Error SignTaskResponse with getting task response header hash. skipping task (this is not expected and should be investigated)", "err", err)
		return nil, nil, err
	}
	msgBytes := taskResponseHash[:]

	sig := config.GetConsensusKeyPair().Sign(msgBytes)

	return sig.Marshal(), data, nil
}

func (k *Keeper) SendSignedTaskResponseToChain(
	ctx context.Context,
	taskId uint64,
	taskResponse []byte,
	blsSignature []byte,
	taskInfo avs.TaskInfo) (string, error) {

	startingEpoch := taskInfo.StartingEpoch
	taskResponsePeriod := taskInfo.TaskResponsePeriod
	taskStatisticalPeriod := taskInfo.TaskStatisticalPeriod

	// Track submission status for each phase
	phaseOneSubmitted := false
	phaseTwoSubmitted := false

	for {
		select {
		case <-ctx.Done():
			k.logger.Info("Shutdown signal received, stopping task response submission", "taskId", taskId)
			return "", ctx.Err() // Gracefully exit if context is canceled
		default:
			// Fetch the current epoch information
			epochIdentifier, err := k.avsReader.GetAVSEpochIdentifier(&bind.CallOpts{}, taskInfo.TaskContractAddress.String())
			if err != nil {
				k.logger.Error("Cannot GetAVSEpochIdentifier", "err", err)
				return "", fmt.Errorf("failed to get AVS info: %w", err) // Stop on persistent error
			}

			num, err := k.avsReader.GetCurrentEpoch(&bind.CallOpts{}, epochIdentifier)
			if err != nil {
				k.logger.Error("Cannot exec GetCurrentEpoch", "err", err)
				return "", fmt.Errorf("failed to get current epoch: %w", err) // Stop on persistent error
			}

			currentEpoch := uint64(num)
			// k.logger.Info("current epoch  is :", "currentEpoch", currentEpoch)
			if currentEpoch > startingEpoch+taskResponsePeriod+taskStatisticalPeriod {
				k.logger.Info("Exiting loop: Task period has passed",
					"Task", taskInfo.TaskContractAddress.String()+"--"+strconv.FormatUint(taskId, 10))
				return "The current task period has passed:", nil
			}

			switch {
			case currentEpoch <= startingEpoch:
				// k.logger.Info("current epoch is less than or equal to the starting epoch", "currentEpoch", currentEpoch, "startingEpoch", startingEpoch, "taskId", taskId)
				time.Sleep(config.GetRetryDelay())

			case currentEpoch <= startingEpoch+taskResponsePeriod:
				if !phaseOneSubmitted {
					k.logger.Info("Execute Phase One Submission Task", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskId", taskId)
					k.logger.Info("Submitting task response for task response period",
						"taskAddr", config.GetAvsGovernanceAddress(), "taskId", taskId, "operator-addr", config.GetKeeperAddress())
					_, err := k.avsWriter.OperatorSubmitTask(
						ctx,
						taskId,
						nil,
						blsSignature,
						config.GetAvsGovernanceAddress(),
						1)
					if err != nil {
						k.logger.Error("Avs failed to OperatorSubmitTask", "err", err, "taskId", taskId)
						return "", fmt.Errorf("failed to submit task during taskResponsePeriod: %w", err)
					}
					phaseOneSubmitted = true
					k.logger.Info("Successfully submitted task response for phase one", "taskId", taskId)
				} else {
					// k.logger.Info("Phase One already submitted", "taskId", taskId)
					time.Sleep(config.GetRetryDelay())
				}

			case currentEpoch <= startingEpoch+taskResponsePeriod+taskStatisticalPeriod && currentEpoch > startingEpoch+taskResponsePeriod:
				if !phaseTwoSubmitted {
					k.logger.Info("Execute Phase Two Submission Task", "currentEpoch", currentEpoch,
						"startingEpoch", startingEpoch, "taskResponsePeriod", taskResponsePeriod, "taskStatisticalPeriod", taskStatisticalPeriod, "taskId", taskId)
					k.logger.Info("Submitting task response for statistical period",
						"taskAddr", config.GetAvsGovernanceAddress(), "taskId", taskId, "operator-addr", config.GetKeeperAddress())
					_, err := k.avsWriter.OperatorSubmitTask(
						ctx,
						taskId,
						taskResponse,
						blsSignature,
						config.GetAvsGovernanceAddress(),
						2)
					if err != nil {
						k.logger.Error("Avs failed to OperatorSubmitTask", "err", err, "taskId", taskId)
						return "", fmt.Errorf("failed to submit task during statistical period: %w", err)
					}
					phaseTwoSubmitted = true
					k.logger.Info("Successfully submitted task response for phase two", "taskId", taskId)
				} else {
					// k.logger.Info("Phase Two already submitted", "taskId", taskId)
					time.Sleep(config.GetRetryDelay())
				}

			default:
				k.logger.Info("Current epoch is not within expected range", "currentEpoch", currentEpoch, "taskId", taskId)
				return "", fmt.Errorf("current epoch %d is not within expected range %d", currentEpoch, startingEpoch)
			}

			// If both phases are submitted, exit the loop
			if phaseOneSubmitted && phaseTwoSubmitted {
				k.logger.Info("Both phases completed successfully", "taskId", taskId)
				return "Both task response phases completed successfully", nil
			}

			// Add a small delay to prevent tight looping, but respect shutdown context
			time.Sleep(config.GetRetryDelay())
		}
	}
}

func (k *Keeper) Close() {
	k.logger.Info("Shutting down keeper...")
	k.cancel() // Cancel the context to signal shutdown to all goroutines
	// k.nodeApi.Stop()
}
