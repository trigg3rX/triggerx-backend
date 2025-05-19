package events

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/client"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/bindings/contractAttestationCenter"
	"github.com/trigg3rX/triggerx-backend/pkg/bindings/contractAvsGovernance"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
)

// retryWithBackoff executes the given function with exponential backoff retry logic
func retryWithBackoff[T any](operation func() (T, error), logger logging.Logger) (T, error) {
	var result T
	var err error
	delay := retryDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err = operation()
		if err == nil {
			return result, nil
		}

		if attempt < maxRetries {
			logger.Warnf("Attempt %d failed: %v. Retrying in %v...", attempt, err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
		}
	}

	return result, fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

func FetchOperatorDetailsAfterDelay(operatorAddress common.Address, delay time.Duration, logger logging.Logger) error {
	logger.Infof("Scheduling fetch of operator details for %s after %v delay", operatorAddress.Hex(), delay)

	go func() {
		time.Sleep(delay)
		logger.Infof("Delay completed, fetching operator details for %s", operatorAddress.Hex())

		err := FetchAndLogOperatorDetails(operatorAddress, logger)
		if err != nil {
			logger.Errorf("Failed to fetch operator details: %v", err)
		}
	}()

	return nil
}

func FetchAndLogOperatorDetails(operatorAddress common.Address, logger logging.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l1Client, err := ethclient.Dial(config.GetEthRPCURL())
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}
	defer l1Client.Close()

	l2Client, err := ethclient.Dial(config.GetBaseRPCURL())
	if err != nil {
		return fmt.Errorf("failed to connect to Base node: %v", err)
	}
	defer l2Client.Close()

	avsGovernanceAddress := common.HexToAddress(config.GetAvsGovernanceAddress())
	avsGovernance, err := contractAvsGovernance.NewAvsGovernanceCaller(avsGovernanceAddress, l1Client)
	if err != nil {
		return fmt.Errorf("failed to create AvsGovernance contract instance: %v", err)
	}

	attestationCenterAddress := common.HexToAddress(config.GetAttestationCenterAddress())
	attestationCenter, err := contractAttestationCenter.NewAttestationCenterCaller(attestationCenterAddress, l2Client)
	if err != nil {
		return fmt.Errorf("failed to create AttestationCenter contract instance: %v", err)
	}

	var restakedStrategiesArr []string
	var rewardsReceiverStr string
	var votingPowerStr string
	var operatorIdStr string

	// Retry getting restaked strategies
	restakedStrategies, err := retryWithBackoff(func() ([]common.Address, error) {
		return avsGovernance.GetOperatorRestakedStrategies(&bind.CallOpts{Context: ctx}, operatorAddress)
	}, logger)
	if err != nil {
		logger.Errorf("Failed to get operator restaked strategies after retries: %v", err)
	} else {
		restakedStrategiesArr = make([]string, len(restakedStrategies))
		for i, strategy := range restakedStrategies {
			restakedStrategiesArr[i] = strategy.Hex()
		}
		logger.Infof("Operator %s restaked strategies: %v", operatorAddress.Hex(), restakedStrategiesArr)
	}

	// Retry getting rewards receiver
	rewardsReceiver, err := retryWithBackoff(func() (common.Address, error) {
		return avsGovernance.GetRewardsReceiver(&bind.CallOpts{Context: ctx}, operatorAddress)
	}, logger)
	if err != nil {
		logger.Errorf("Failed to get rewards receiver after retries: %v", err)
	} else {
		rewardsReceiverStr = rewardsReceiver.Hex()
		logger.Infof("Operator %s rewards receiver: %s", operatorAddress.Hex(), rewardsReceiverStr)
	}

	// Retry getting L2 voting power
	l2VotingPower, err := retryWithBackoff(func() (*big.Int, error) {
		return attestationCenter.VotingPower(&bind.CallOpts{Context: ctx}, operatorAddress)
	}, logger)
	if err != nil {
		logger.Errorf("Failed to get L2 voting power after retries: %v", err)
	} else {
		votingPowerStr = l2VotingPower.String()
		logger.Infof("Operator %s L2 voting power: %s", operatorAddress.Hex(), votingPowerStr)
	}

	// Retry getting operator ID
	operatorId, err := retryWithBackoff(func() (*big.Int, error) {
		return attestationCenter.OperatorsIdsByAddress(&bind.CallOpts{Context: ctx}, operatorAddress)
	}, logger)
	if err != nil {
		logger.Errorf("Failed to get operator ID after retries: %v", err)
	} else {
		operatorIdStr = operatorId.String()
		logger.Infof("Operator %s ID: %s", operatorAddress.Hex(), operatorIdStr)
	}

	if operatorIdStr != "" && votingPowerStr != "" {
		err = client.UpdateOperatorDetails(
			operatorAddress.Hex(),
			operatorIdStr,
			votingPowerStr,
			rewardsReceiverStr,
			restakedStrategiesArr,
		)
		if err != nil {
			logger.Errorf("Failed to update operator details in database: %v", err)
		} else {
			logger.Infof("Successfully updated operator details in database for %s", operatorAddress.Hex())
		}
	} else {
		logger.Errorf("Missing required data for operator %s, skipping database update", operatorAddress.Hex())
	}

	return nil
}
