package quorum

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	regcoord "github.com/trigg3rX/triggerx-backend/pkg/avsinterface/bindings/RegistryCoordinator"
)

const (
	MAX_OPERATORS_PER_QUORUM = 50
	TOTAL_QUORUMS            = 5
	CONTRACT_ADDRESS         = "0x13a05d12b8061f8F12beCa62a42b981531021439"
	HOLESKY_RPC              = "https://ethereum-holesky-rpc.publicnode.com/"
)

func Create() error {
	log.Println("Creating quorum...")
	client, err := ethclient.Dial(HOLESKY_RPC)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	contract, err := regcoord.NewContractRegistryCoordinator(common.HexToAddress(CONTRACT_ADDRESS), client)
	if err != nil {
		log.Fatalf("Failed to create contract: %v", err)
	}

	count, err := contract.QuorumCount(nil)
	if err != nil {
		log.Fatalf("Failed to get quorum count: %v", err)
	}
	fmt.Printf("Current number of quorums: %v\n", count)

	auth, err := createAuthTransactor()
	if err != nil {
		log.Fatalf("failed to create auth transactor: %v", err)
	}

	// Define operator set parameters
	operatorSetParams := regcoord.IRegistryCoordinatorOperatorSetParam{
		MaxOperatorCount:        uint32(MAX_OPERATORS_PER_QUORUM),
		KickBIPsOfOperatorStake: uint16(100), // 1% in basis points (100 = 1%)
		KickBIPsOfTotalStake:    uint16(200), // 2% in basis points (200 = 2%)
	}

	// Define minimum stake (example: 32 ETH in wei)
	minimumStake := big.NewInt(0)
	minimumStake.SetString("32000000000000000000", 10) // 32 ETH in wei

	// Define strategy parameters
	// This is an example - adjust according to your needs
	strategyParams := []regcoord.IStakeRegistryStrategyParams{
		{
			Strategy:   common.HexToAddress("0x..."), // Add strategy contract address
			Multiplier: big.NewInt(1000),             // Multiplier in basis points (1000 = 10%)
		},
		// Add more strategies if needed
	}

	// Create the quorum
	tx, err := contract.CreateQuorum(
		auth,
		operatorSetParams,
		minimumStake,
		strategyParams,
	)
	if err != nil {
		log.Fatalf("failed to create quorum: %v", err)
	}

	log.Printf("Quorum creation transaction submitted: %s", tx.Hash().Hex())

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("failed to wait for quorum creation: %v", err)
	}

	if receipt.Status == 0 {
		log.Fatalf("quorum creation transaction failed")
	}

	// Get updated quorum count
	newCount, err := contract.QuorumCount(nil)
	if err != nil {
		log.Fatalf("failed to get updated quorum count: %v", err)
	}
	log.Printf("New number of quorums: %v\n", newCount)

	return nil
}

// RegisterOperator registers an operator for specific quorums
func RegisterOperator(quorumNumbers []byte, socket string,
	pubkeyParams regcoord.IBLSApkRegistryPubkeyRegistrationParams,
	operatorSignature regcoord.ISignatureUtilsSignatureWithSaltAndExpiry) error {
	client, err := ethclient.Dial(HOLESKY_RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	contract, err := regcoord.NewContractRegistryCoordinator(common.HexToAddress(CONTRACT_ADDRESS), client)
	if err != nil {
		return fmt.Errorf("failed to create contract instance: %v", err)
	}

	// Create auth transaction options (you'll need to add your private key)
	auth, err := createAuthTransactor()
	if err != nil {
		return fmt.Errorf("failed to create auth transactor: %v", err)
	}

	tx, err := contract.RegisterOperator(auth, quorumNumbers, socket, pubkeyParams, operatorSignature)
	if err != nil {
		return fmt.Errorf("failed to register operator: %v", err)
	}

	log.Printf("Operator registration transaction submitted: %s", tx.Hash().Hex())
	return nil
}

func DeregisterOperator(quorumNumbers []byte) error {
	client, err := ethclient.Dial(HOLESKY_RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	contract, err := regcoord.NewContractRegistryCoordinator(common.HexToAddress(CONTRACT_ADDRESS), client)
	if err != nil {
		return fmt.Errorf("failed to create contract instance: %v", err)
	}

	auth, err := createAuthTransactor()
	if err != nil {
		return fmt.Errorf("failed to create auth transactor: %v", err)
	}

	tx, err := contract.DeregisterOperator(auth, quorumNumbers)
	if err != nil {
		return fmt.Errorf("failed to deregister operator: %v", err)
	}

	log.Printf("Operator deregistration transaction submitted: %s", tx.Hash().Hex())
	return nil
}

func createAuthTransactor() (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA("your-private-key-here")
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Create client first
	client, err := ethclient.Dial(HOLESKY_RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	// Get chainID from client
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	return auth, nil
}
