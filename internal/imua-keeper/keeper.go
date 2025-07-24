package keeper

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
	avs "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
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

	return &Keeper{
		logger:      logger,
		ethClient:   ethClient,
		ethWsClient: ethWsClient,
		nodeApi:     nodeApi,
		avsReader:   avsReader,
		avsWriter:   avsWriter,
	}
}

func (k *Keeper) Start(ctx context.Context) error {
	k.logger.Infof("Starting operator.")
	k.nodeApi.Start()

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
				sig, resBytes, err := k.SignTaskResponse()
				if err != nil {
					k.logger.Error("Failed to sign task response", "err", err)
					continue
				}
				taskInfo, _ := k.avsReader.GetTaskInfo(&bind.CallOpts{}, config.GetAvsGovernanceAddress(), e.TaskID)
				go func() {
					_, err := k.SendSignedTaskResponseToChain(context.Background(), e.TaskID, resBytes, sig, taskInfo)
					if err != nil {
						k.logger.Error("Failed to send signed task response to chain", "err", err)
					}
				}()
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
		"taskID", e.TaskID,
		"definitionHash", fmt.Sprintf("0x%x", e.DefinitionHash),
		"kind", e.Kind)

	// TODO: Implement task processing logic
	// This could include:
	// - Fetching task details from the contract
	// - Preparing task execution data
	// - Setting up monitoring for the task
}

func (k *Keeper) SignTaskResponse() ([]byte, []byte, error) {
	// TODO: Implement task response signing
	// This should:
	// 1. Create a task response structure
	// 2. Sign it with the BLS key
	// 3. Return the signature and response bytes

	// For now, return placeholder values
	responseBytes := []byte("task_response_placeholder")
	signature := []byte("bls_signature_placeholder")

	return signature, responseBytes, nil
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
				k.logger.Info("current epoch is less than or equal to the starting epoch", "currentEpoch", currentEpoch, "startingEpoch", startingEpoch, "taskId", taskId)
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
					k.logger.Info("Phase One already submitted", "taskId", taskId)
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
					k.logger.Info("Phase Two already submitted", "taskId", taskId)
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

			// Add a small delay to prevent tight looping
			time.Sleep(config.GetRetryDelay())
		}
	}
}
