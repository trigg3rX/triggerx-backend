package main

import (
	"fmt"
	"log"
	"math/big"
	"time"
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/go-backend/pkg/database"
	"strings"
)

const (
	MAX_KEEPERS_PER_QUORUM = 50
	TOTAL_QUORUMS          = 5
	CONTRACT_ADDRESS       = "0x13a05d12b8061f8F12beCa62a42b981531021439"
	HOLESKY_RPC            = "https://ethereum-holesky-rpc.publicnode.com/"
)

// Contract ABI as a string
const ContractABI = `[{"inputs":[{"internalType":"bytes","name":"operatorId","type":"bytes"},{"internalType":"string","name":"socket","type":"string"},{"components":[{"components":[{"internalType":"uint256","name":"X","type":"uint256"},{"internalType":"uint256","name":"Y","type":"uint256"}],"internalType":"struct BN254.G1Point","name":"pubkeyG1","type":"tuple"},{"components":[{"internalType":"uint256[2]","name":"X","type":"uint256[2]"},{"internalType":"uint256[2]","name":"Y","type":"uint256[2]"}],"internalType":"struct BN254.G2Point","name":"pubkeyG2","type":"tuple"}],"internalType":"struct IBLSApkRegistry.PubkeyRegistrationParams","name":"pubkeyRegistrationParams","type":"tuple"},{"components":[{"internalType":"bytes","name":"signature","type":"bytes"},{"internalType":"bytes32","name":"salt","type":"bytes32"},{"internalType":"uint256","name":"expiry","type":"uint256"}],"internalType":"struct ISignatureUtils.SignatureWithSaltAndExpiry","name":"operatorSignature","type":"tuple"}],"name":"registerOperator","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"}],"name":"deregisterOperator","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

type QuorumManager struct {
	db       *database.Connection
	client   *ethclient.Client
	contract *bind.BoundContract
}

func NewQuorumManager() (*QuorumManager, error) {
	// Initialize database connection
	config := database.NewConfig()
	conn, err := database.NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Connect to Ethereum client
	client, err := ethclient.Dial(HOLESKY_RPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Create bound contract
	contractAddress := common.HexToAddress(CONTRACT_ADDRESS)
	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	return &QuorumManager{
		db:       conn,
		client:   client,
		contract: contract,
	}, nil
}

func (qm *QuorumManager) GetQuorumNo(keeperID int64) (int, error) {
	var quorumNo int
	err := qm.db.Session().Query(`
		SELECT current_quorum_no 
		FROM triggerx.keeper_data 
		WHERE keeper_id = ?`, keeperID).Scan(&quorumNo)

	if err != nil {
		return 0, fmt.Errorf("failed to get quorum number: %v", err)
	}
	return quorumNo, nil
}

func (qm *QuorumManager) RegisterKeeper(keeperID int64, operatorAddress string, privateKey string) error {
	// Create auth transaction options
	auth, err := createTransactOpts(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Convert operator address to bytes
	operatorBytes := common.HexToAddress(operatorAddress).Bytes()

	// Create socket value (must be non-empty)
	socket := "socket://localhost:8545"
	

	// Initialize G1 point
	pubkeyG1 := struct {
		X *big.Int
		Y *big.Int
	}{
		X: new(big.Int).SetInt64(1),
		Y: new(big.Int).SetInt64(2),
	}

	// Initialize G2 point
	pubkeyG2 := struct {
		X [2]*big.Int
		Y [2]*big.Int
	}{
		X: [2]*big.Int{new(big.Int).SetInt64(3), new(big.Int).SetInt64(4)},
		Y: [2]*big.Int{new(big.Int).SetInt64(5), new(big.Int).SetInt64(6)},
	}

	// Create pubkey registration
	pubkeyRegistration := struct {
		PubkeyG1 struct {
			X *big.Int
			Y *big.Int
		}
		PubkeyG2 struct {
			X [2]*big.Int
			Y [2]*big.Int
		}
	}{
		PubkeyG1: pubkeyG1,
		PubkeyG2: pubkeyG2,
	}

	// Create signature
	var salt [32]byte
	copy(salt[:], []byte("random_salt_value_12345678901234567"))

	dummySignature := make([]byte, 96)
	for i := range dummySignature {
		dummySignature[i] = byte(i % 256)
	}

	signature := struct {
		Signature []byte
		Salt      [32]byte
		Expiry    *big.Int
	}{
		Signature: dummySignature,
		Salt:      salt,
		Expiry:    new(big.Int).SetInt64(time.Now().Add(24 * time.Hour).Unix()),
	}

	// Call contract register operator function
	tx, err := qm.contract.Transact(auth, "registerOperator", 
		operatorBytes,
		socket,
		pubkeyRegistration,
		signature,
	)
	if err != nil {
		return fmt.Errorf("failed to register operator: %v", err)
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), qm.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	// Get available quorum
	quorumNo, err := qm.GetAvailableQuorum()
	if err != nil {
		return err
	}

	// Store in database
	err = qm.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET current_quorum_no = ?, register_tx_hash = ?
		WHERE keeper_id = ?`,
		quorumNo, tx.Hash().String(), keeperID).Exec()

	if err != nil {
		return fmt.Errorf("failed to update keeper data: %v", err)
	}

	return nil
}

func (qm *QuorumManager) DeregisterKeeper(keeperID int64, operatorAddress string, privateKey string) error {
	// Create auth transaction options
	auth, err := createTransactOpts(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Convert operator address to bytes
	operatorBytes := common.HexToAddress(operatorAddress).Bytes()

	// Call contract deregister operator function
	tx, err := qm.contract.Transact(auth, "deregisterOperator", operatorBytes)
	if err != nil {
		return fmt.Errorf("failed to deregister operator: %v", err)
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), qm.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	// Remove from quorum in database
	err = qm.db.Session().Query(`
		UPDATE triggerx.keeper_data 
		SET current_quorum_no = 0
		WHERE keeper_id = ?`, keeperID).Exec()

	if err != nil {
		return fmt.Errorf("failed to update keeper data: %v", err)
	}

	return nil
}

func (qm *QuorumManager) GetAvailableQuorum() (int, error) {
	// Query each quorum to find one with space
	for quorumNo := 1; quorumNo <= TOTAL_QUORUMS; quorumNo++ {
		var count int
		err := qm.db.Session().Query(`
			SELECT COUNT(*) 
			FROM triggerx.keeper_data 
			WHERE current_quorum_no = ?`, quorumNo).Scan(&count)

		if err != nil {
			return 0, fmt.Errorf("failed to get quorum count: %v", err)
		}

		if count < MAX_KEEPERS_PER_QUORUM {
			return quorumNo, nil
		}
	}

	return -1, fmt.Errorf("no available quorums")
}

// Helper function to create transaction options
func createTransactOpts(privateKeyHex string) (*bind.TransactOpts, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(17000)) // 17000 is Holesky chain ID
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	return auth, nil
}

func main() {
	fmt.Println("Initializing Quorum Creator...")

	qm, err := NewQuorumManager()
	if err != nil {
		log.Fatalf("Failed to initialize quorum manager: %v", err)
	}
	defer qm.db.Close()

	// Example usage
	keeperID := int64(2)
	operatorAddress := "0xC76EA60887CA82C474cf6dfc17f918DDd68D6cA2"                  // Replace with actual address
	privateKey := "29b47d5446e76cdfc0fb55cfbddea308f3d5f0c4151105f39d262f7dd49e9600" // Replace with actual private key (without 0x prefix)

	// Example: Register a keeper
	fmt.Println("Attempting to register keeper...")
	err = qm.RegisterKeeper(keeperID, operatorAddress, privateKey)
	if err != nil {
		log.Printf("Failed to register keeper: %v", err)
	} else {
		log.Printf("Successfully registered keeper %d", keeperID)
	}

	// Example: Get keeper's quorum number
	quorumNo, err := qm.GetQuorumNo(keeperID)
	if err != nil {
		log.Printf("Failed to get quorum number: %v", err)
	} else {
		log.Printf("Keeper %d is in quorum %d", keeperID, quorumNo)
	}
}
