package execution

import (
	// "context"
	// "fmt"
	// "math/big"
	// "strings"

	// "github.com/ethereum/go-ethereum/accounts/abi"
	// ethcommon "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/types"
	// "github.com/ethereum/go-ethereum/crypto"
	// "github.com/trigg3rX/triggerx-backend/internal/keeper/config"
)


// func (e *JobExecutor) prepareTransaction(contractAddress ethcommon.Address, callData []byte) (*types.Transaction, error) {
// 	// privateKey, err := crypto.HexToECDSA(config.PrivateKeyController)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("failed to parse private key: %v", err)
// 	// }

// 	nonce, err := e.ethClient.PendingNonceAt(context.Background(), ethcommon.HexToAddress(config.KeeperAddress))
// 	if err != nil {
// 		return nil, err
// 	}

// 	gasPrice, err := e.ethClient.SuggestGasPrice(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}

// 	executionABI, err := abi.JSON(strings.NewReader(executionABIJSON))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse execution contract ABI: %v", err)
// 	}

// 	executionInput, err := executionABI.Pack("executeFunction", contractAddress, callData)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to pack execution contract input: %v", err)
// 	}

// 	return types.NewTransaction(nonce, ethcommon.HexToAddress(executionContractAddress), big.NewInt(0), 300000, gasPrice, executionInput), nil
// }

// const executionABIJSON = `[{"inputs":[{"name":"target","type":"address"},{"name":"data","type":"bytes"}],"name":"executeFunction","outputs":[],"stateMutability":"payable","type":"function"}]`
