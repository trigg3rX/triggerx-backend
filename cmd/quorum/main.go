package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/trigg3rX/go-backend/pkg/database"
)

const (
	MAX_OPERATORS_PER_QUORUM = 50
	TOTAL_QUORUMS            = 5
	CONTRACT_ADDRESS         = "0x13a05d12b8061f8F12beCa62a42b981531021439"
	HOLESKY_RPC              = "https://ethereum-holesky-rpc.publicnode.com/"
)

// Contract ABI as a string
const ContractABI = `[
    {
        "inputs": [
            {"internalType": "bytes", "name": "quorumNumbers", "type": "bytes"},
            {"internalType": "string", "name": "socket", "type": "string"},
            {
                "components": [
                    {
                        "components": [
                            {"internalType": "uint256", "name": "X", "type": "uint256"},
                            {"internalType": "uint256", "name": "Y", "type": "uint256"}
                        ],
                        "internalType": "struct BN254.G1Point",
                        "name": "pubkeyG1",
                        "type": "tuple"
                    },
                    {
                        "components": [
                            {"internalType": "uint256[2]", "name": "X", "type": "uint256[2]"},
                            {"internalType": "uint256[2]", "name": "Y", "type": "uint256[2]"}
                        ],
                        "internalType": "struct BN254.G2Point",
                        "name": "pubkeyG2",
                        "type": "tuple"
                    }
                ],
                "internalType": "struct IBLSApkRegistry.PubkeyRegistrationParams",
                "name": "params",
                "type": "tuple"
            },
            {
                "components": [
                    {"internalType": "bytes", "name": "signature", "type": "bytes"},
                    {"internalType": "bytes32", "name": "salt", "type": "bytes32"},
                    {"internalType": "uint256", "name": "expiry", "type": "uint256"}
                ],
                "internalType": "struct ISignatureUtils.SignatureWithSaltAndExpiry",
                "name": "operatorSignature",
                "type": "tuple"
            }
        ],
        "name": "registerOperator",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"internalType": "address", "name": "operator", "type": "address"}
        ],
        "name": "deregisterOperator",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]`

// QuorumManager manages operator registration and quorum operations
type QuorumManager struct {
	db       *database.Connection
	client   *ethclient.Client
	contract *bind.BoundContract
}

// PubkeyRegistrationParams represents the registration parameters for an operator's public key
type PubkeyRegistrationParams struct {
	PubkeyG1 struct {
		X *big.Int
		Y *big.Int
	}
	PubkeyG2 struct {
		X [2]*big.Int
		Y [2]*big.Int
	}
}

// SignatureWithSaltAndExpiry represents the operator's signature for registration
type SignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// NewQuorumManager initializes and returns a new QuorumManager
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

// RegisterOperator registers an operator for specified quorums
func (qm *QuorumManager) RegisterOperator(
	quorumNumbers []byte,
	socket string,
	params PubkeyRegistrationParams,
	operatorSignature SignatureWithSaltAndExpiry,
	privateKey string,
) error {
	// Create auth transaction options
	auth, err := createTransactOpts(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Call contract register operator function
	tx, err := qm.contract.Transact(auth, "registerOperator",
		quorumNumbers,
		socket,
		params,
		operatorSignature,
	)
	if err != nil {
		return fmt.Errorf("failed to register operator: %v", err)
	}

	// Wait for transaction to be mined
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	receipt, err := bind.WaitMined(ctx, qm.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	// Validate quorum constraints
	for _, quorumNumber := range quorumNumbers {
		keeperCount, err := qm.getOperatorCountForQuorum(int(quorumNumber))
		if err != nil {
			return fmt.Errorf("failed to check quorum count: %v", err)
		}

		if keeperCount > MAX_OPERATORS_PER_QUORUM {
			return fmt.Errorf("operator count exceeds maximum for quorum %d", quorumNumber)
		}
	}

	return nil
}

// DeregisterOperator removes an operator from the system
func (qm *QuorumManager) DeregisterOperator(
	operatorAddress common.Address,
	privateKey string,
) error {
	// Create auth transaction options
	auth, err := createTransactOpts(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create transaction options: %v", err)
	}

	// Call contract deregister operator function
	tx, err := qm.contract.Transact(auth, "deregisterOperator", operatorAddress)
	if err != nil {
		return fmt.Errorf("failed to deregister operator: %v", err)
	}

	// Wait for transaction to be mined
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	receipt, err := bind.WaitMined(ctx, qm.client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	return nil
}

// GetOperatorQuorum retrieves the current quorum for a specific operator
func (qm *QuorumManager) GetOperatorQuorum(operatorID int64) (int, error) {
	var quorumNo int
	err := qm.db.Session().Query(`
		SELECT current_quorum_no 
		FROM triggerx.operator_data 
		WHERE operator_id = ?`, operatorID).Scan(&quorumNo)

	if err != nil {
		return 0, fmt.Errorf("failed to get quorum number: %v", err)
	}
	return quorumNo, nil
}

// GetAvailableQuorum finds a quorum with available capacity
func (qm *QuorumManager) GetAvailableQuorum() (int, error) {
	for quorumNo := 1; quorumNo <= TOTAL_QUORUMS; quorumNo++ {
		count, err := qm.getOperatorCountForQuorum(quorumNo)
		if err != nil {
			return 0, fmt.Errorf("failed to get quorum count: %v", err)
		}

		if count < MAX_OPERATORS_PER_QUORUM {
			return quorumNo, nil
		}
	}

	return -1, fmt.Errorf("no available quorums")
}

// getOperatorCountForQuorum counts the number of operators in a specific quorum
func (qm *QuorumManager) getOperatorCountForQuorum(quorumNumber int) (int, error) {
	var count int
	err := qm.db.Session().Query(`
		SELECT COUNT(*) 
		FROM triggerx.operator_data 
		WHERE current_quorum_no = ?`, quorumNumber).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get quorum count: %v", err)
	}

	return count, nil
}

// createTransactOpts creates transaction options from a private key
func createTransactOpts(privateKeyHex string) (*bind.TransactOpts, error) {
	// Remove '0x' prefix if present
	if strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = privateKeyHex[2:]
	}

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
	fmt.Println("Initializing Operator Registration...")

	// Create QuorumManager
	qm, err := NewQuorumManager()
	if err != nil {
		log.Fatalf("Failed to initialize quorum manager: %v", err)
	}
	defer qm.db.Close()

	// Prepare registration parameters
	quorumNumbers := []byte{1, 2} // Example quorum numbers
	socket := "socket://localhost:8545"

	// Dummy pubkey registration params
	params := PubkeyRegistrationParams{
		PubkeyG1: struct {
			X *big.Int
			Y *big.Int
		}{
			X: new(big.Int).SetInt64(1),
			Y: new(big.Int).SetInt64(2),
		},
		PubkeyG2: struct {
			X [2]*big.Int
			Y [2]*big.Int
		}{
			X: [2]*big.Int{new(big.Int).SetInt64(3), new(big.Int).SetInt64(4)},
			Y: [2]*big.Int{new(big.Int).SetInt64(5), new(big.Int).SetInt64(6)},
		},
	}

	// Prepare operator signature
	var salt [32]byte
	copy(salt[:], []byte("random_salt_value_12345678901234567"))

	dummySignature := make([]byte, 96)
	for i := range dummySignature {
		dummySignature[i] = byte(i % 256)
	}

	operatorSignature := SignatureWithSaltAndExpiry{
		Signature: dummySignature,
		Salt:      salt,
		Expiry:    new(big.Int).SetInt64(time.Now().Add(24 * time.Hour).Unix()),
	}

	// Private key for transaction (replace with actual private key)
	privateKey := ""



	
	// Attempt to register operator
	err = qm.RegisterOperator(quorumNumbers, socket, params, operatorSignature, privateKey)
	if err != nil {
		log.Printf("Failed to register operator: %v", err)
	} else {
		log.Println("Successfully registered operator")
	}
}
