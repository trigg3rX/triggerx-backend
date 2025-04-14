package registrar

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/internal/registrar/database"
	"github.com/trigg3rX/triggerx-backend/pkg/bindings/contractAttestationCenter"
	"github.com/trigg3rX/triggerx-backend/pkg/bindings/contractAvsGovernance"
)

// FetchOperatorDetailsAfterDelay schedules fetching operator details after a specified delay
func FetchOperatorDetailsAfterDelay(operatorAddress common.Address, delay time.Duration) {
	// Log that we're scheduling the fetch
	logger.Infof("Scheduling fetch of operator details for %s after %v delay", operatorAddress.Hex(), delay)

	// Start a goroutine to wait for the specified delay and then fetch the details
	go func() {
		time.Sleep(delay)
		logger.Infof("Delay completed, fetching operator details for %s", operatorAddress.Hex())

		err := FetchAndLogOperatorDetails(operatorAddress)
		if err != nil {
			logger.Errorf("Failed to fetch operator details: %v", err)
		}
	}()
}

// FetchAndLogOperatorDetails fetches operator details from both contracts and logs them
func FetchAndLogOperatorDetails(operatorAddress common.Address) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to L1 and L2 nodes
	l1Client, err := ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}
	defer l1Client.Close()

	l2Client, err := ethclient.Dial(config.BaseRpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to Base node: %v", err)
	}
	defer l2Client.Close()

	// Create AvsGovernance contract instance
	avsGovernanceAddress := common.HexToAddress(config.AvsGovernanceAddress)
	avsGovernance, err := contractAvsGovernance.NewAvsGovernanceCaller(avsGovernanceAddress, l1Client)
	if err != nil {
		return fmt.Errorf("failed to create AvsGovernance contract instance: %v", err)
	}

	// Create AttestationCenter contract instance
	attestationCenterAddress := common.HexToAddress(config.AttestationCenterAddress)
	attestationCenter, err := contractAttestationCenter.NewAttestationCenterCaller(attestationCenterAddress, l2Client)
	if err != nil {
		return fmt.Errorf("failed to create AttestationCenter contract instance: %v", err)
	}

	// Variables to store the contract data
	var restakedStrategiesArr []string
	var rewardsReceiverStr string
	var votingPowerStr string
	var operatorIdStr string

	// Fetch restakedStrategies from AvsGovernance (L1)
	restakedStrategies, err := avsGovernance.GetOperatorRestakedStrategies(&bind.CallOpts{Context: ctx}, operatorAddress)
	if err != nil {
		logger.Errorf("Failed to get operator restaked strategies: %v", err)
	} else {
		// Convert to string array for logging and storage
		restakedStrategiesArr = make([]string, len(restakedStrategies))
		for i, strategy := range restakedStrategies {
			restakedStrategiesArr[i] = strategy.Hex()
		}
		logger.Infof("Operator %s restaked strategies: %v", operatorAddress.Hex(), restakedStrategiesArr)
	}

	// Fetch rewardsReceiver from AvsGovernance (L1)
	rewardsReceiver, err := avsGovernance.GetRewardsReceiver(&bind.CallOpts{Context: ctx}, operatorAddress)
	if err != nil {
		logger.Errorf("Failed to get rewards receiver: %v", err)
	} else {
		rewardsReceiverStr = rewardsReceiver.Hex()
		logger.Infof("Operator %s rewards receiver: %s", operatorAddress.Hex(), rewardsReceiverStr)
	}

	// Fetch voting power from AttestationCenter (L2)
	l2VotingPower, err := attestationCenter.VotingPower(&bind.CallOpts{Context: ctx}, operatorAddress)
	if err != nil {
		logger.Errorf("Failed to get L2 voting power: %v", err)
	} else {
		votingPowerStr = l2VotingPower.String()
		logger.Infof("Operator %s L2 voting power: %s", operatorAddress.Hex(), votingPowerStr)
	}

	// Fetch operatorsIdsByAddress from AttestationCenter (L2)
	operatorId, err := attestationCenter.OperatorsIdsByAddress(&bind.CallOpts{Context: ctx}, operatorAddress)
	if err != nil {
		logger.Errorf("Failed to get operator ID: %v", err)
	} else {
		operatorIdStr = operatorId.String()
		logger.Infof("Operator %s ID: %s", operatorAddress.Hex(), operatorIdStr)
	}

	// Update database with the fetched information
	if operatorIdStr != "" && votingPowerStr != "" {
		err = database.UpdateOperatorDetails(
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

// TestFetchOperatorDetails is a utility function to manually trigger operator details fetching
// It can be used for testing or debugging
func TestFetchOperatorDetails(operatorAddressHex string) error {
	operatorAddress := common.HexToAddress(operatorAddressHex)
	logger.Infof("Manually triggering fetch of operator details for %s", operatorAddress.Hex())
	return FetchAndLogOperatorDetails(operatorAddress)
}
