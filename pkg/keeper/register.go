package keeper

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v2"

	// "github.com/Layr-Labs/eigensdk-go/crypto/bls"
	eigenSdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	// regcoord "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
)

type OperatorConfig struct {
	Environment struct {
		EthRpcUrl string `yaml:"ethrpcurl"`
		EthWsUrl  string `yaml:"ethwsurl"`
	} `yaml:"environment"`
	AVS_NAME  string `yaml:"avs_name"`
	SEM_VER   string `yaml:"sem_ver"`
	Addresses struct {
		ServiceManagerAddress  string `yaml:"service_manager_address"`
		OperatorStateRetriever string `yaml:"operator_state_retriever"`
	} `yaml:"addresses"`
	Prometheus struct {
		PortAddress string `yaml:"port_address"`
	} `yaml:"prometheus"`
}

func loadConfig(path string) (*OperatorConfig, error) {
	config := &OperatorConfig{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate required fields
	if config.AVS_NAME == "" {
		return nil, fmt.Errorf("avs_name is required in configuration")
	}
	if config.SEM_VER == "" {
		return nil, fmt.Errorf("sem_ver is required in configuration")
	}
	if config.Environment.EthRpcUrl == "" {
		return nil, fmt.Errorf("ethrpcurl is required in configuration")
	}

	return config, nil
}

func (o *Keeper) DepositIntoStrategy(strategyAddr common.Address, amount *big.Int) error {
	_, tokenAddr, err := o.EigenlayerReader.GetStrategyAndUnderlyingToken(context.Background(), strategyAddr)
	if err != nil {
		o.Logger.Error("Failed to fetch strategy contract", "err", err)
		return err
	}
	contractErc20Mock, err := o.AvsReader.GetErc20Mock(context.Background(), tokenAddr)
	if err != nil {
		o.Logger.Error("Failed to fetch ERC20Mock contract", "err", err)
		return err
	}
	txOpts, err := o.AvsWriter.GetTxMgr().GetNoSendTxOpts()
	if err != nil {
		o.Logger.Errorf("Error getting no send tx opts")
		return err
	}
	tx, err := contractErc20Mock.Mint(txOpts, o.KeeperAddr, amount)
	if err != nil {
		o.Logger.Errorf("Error assembling Mint tx")
		return err
	}
	_, err = o.AvsWriter.GetTxMgr().Send(context.Background(), tx, true)
	if err != nil {
		o.Logger.Errorf("Error submitting Mint tx")
		return err
	}

	_, err = o.EigenlayerWriter.DepositERC20IntoStrategy(context.Background(), strategyAddr, amount, true)
	if err != nil {
		o.Logger.Errorf("Error depositing into strategy", "err", err)
		return err
	}
	return nil
}

// Registration specific functions
func (o *Keeper) RegisterOperatorWithAvs(operatorEcdsaKeyPair *ecdsa.PrivateKey) error {
	// Load the config file
	config, err := loadConfig("triggerx_operator.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create a new eth client with the URL from config
	ethClient, err := ethclient.Dial(config.Environment.EthRpcUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}
	o.EthClient = ethClient

	// 1. First check if operator is already registered
	operatorId, err := o.AvsReader.GetOperatorId(&bind.CallOpts{}, o.KeeperAddr)
	if err != nil {
		return fmt.Errorf("failed to check operator registration: %w", err)
	}
	if operatorId != [32]byte{} {
		o.Logger.Info("Operator already registered with AVS")
		return nil
	}

	// 2. Set up registration parameters
	quorumNumbers := eigenSdkTypes.QuorumNums{eigenSdkTypes.QuorumNum(0)} // Default quorum
	socket := "Not Needed"                                                // As specified in your code

	// Generate a random salt for registration
	operatorToAvsRegistrationSigSalt := [32]byte{123} // Consider using crypto/rand for production

	// 3. Get current block for expiry calculation
	curBlockNum, err := o.EthClient.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("unable to get current block number: %w", err)
	}

	curBlock, err := o.EthClient.HeaderByNumber(context.Background(), big.NewInt(int64(curBlockNum)))
	if err != nil {
		return fmt.Errorf("unable to get current block: %w", err)
	}

	// Set signature validity period (about 11.5 days)
	sigValidForSeconds := int64(1_000_000)
	operatorToAvsRegistrationSigExpiry := big.NewInt(int64(curBlock.Time) + sigValidForSeconds)

	// 4. Register operator with AVS
	receipt, err := o.AvsWriter.RegisterOperatorInQuorumWithAVSRegistryCoordinator(
		context.Background(),
		operatorEcdsaKeyPair,
		operatorToAvsRegistrationSigSalt,
		operatorToAvsRegistrationSigExpiry,
		o.BlsKeypair,
		quorumNumbers,
		socket,
		true, // Wait for confirmation
	)
	if err != nil {
		return fmt.Errorf("failed to register operator with AVS registry coordinator: %w", err)
	}

	o.Logger.Info("Successfully registered operator with AVS",
		"txHash", receipt.TxHash.Hex(),
		"blockNumber", receipt.BlockNumber,
		"gasUsed", receipt.GasUsed,
	)

	// 5. Verify registration
	newOperatorId, err := o.AvsReader.GetOperatorId(&bind.CallOpts{}, o.KeeperAddr)
	if err != nil {
		return fmt.Errorf("failed to verify operator registration: %w", err)
	}
	if newOperatorId == [32]byte{} {
		return fmt.Errorf("operator registration failed: operator ID is still empty after registration")
	}

	// Store the operator ID
	o.KeeperId = newOperatorId

	o.Logger.Info("Registration verified successfully",
		"operatorId", hex.EncodeToString(newOperatorId[:]),
		"address", o.KeeperAddr.Hex(),
	)

	return nil
}

// PRINTING STATUS OF OPERATOR: 1
// operator address: 0xa0ee7a142d267c1f36714e4a8f75612f20a79720
// dummy token balance: 0
// delegated shares in dummyTokenStrat: 200
// operator pubkey hash in AVS pubkey compendium (0 if not registered): 0x4b7b8243d970ff1c90a7c775c008baad825893ec6e806dfa5d3663dc093ed17f
// operator is opted in to eigenlayer: true
// operator is opted in to playgroundAVS (aka can be slashed): true
// operator status in AVS registry: REGISTERED
//
//	operatorId: 0x4b7b8243d970ff1c90a7c775c008baad825893ec6e806dfa5d3663dc093ed17f
//	middlewareTimesLen (# of stake updates): 0
//
// operator is frozen: false
type OperatorStatus struct {
	EcdsaAddress string
	// pubkey compendium related
	PubkeysRegistered bool
	G1Pubkey          string
	G2Pubkey          string
	// avs related
	RegisteredWithAvs bool
	OperatorId        string
}

func (o *Keeper) PrintOperatorStatus() error {
	fmt.Println("Printing operator status")
	operatorId, err := o.AvsReader.GetOperatorId(&bind.CallOpts{}, o.KeeperAddr)
	if err != nil {
		return err
	}
	pubkeysRegistered := operatorId != [32]byte{}
	registeredWithAvs := o.KeeperId != [32]byte{}
	operatorStatus := OperatorStatus{
		EcdsaAddress:      o.KeeperAddr.String(),
		PubkeysRegistered: pubkeysRegistered,
		G1Pubkey:          o.BlsKeypair.GetPubKeyG1().String(),
		G2Pubkey:          o.BlsKeypair.GetPubKeyG2().String(),
		RegisteredWithAvs: registeredWithAvs,
		OperatorId:        hex.EncodeToString(o.KeeperId[:]),
	}
	operatorStatusJson, err := json.MarshalIndent(operatorStatus, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(operatorStatusJson))
	return nil
}

// func pubKeyG1ToBN254G1Point(p *bls.G1Point) regcoord.BN254G1Point {
// 	return regcoord.BN254G1Point{
// 		X: p.X.BigInt(new(big.Int)),
// 		Y: p.Y.BigInt(new(big.Int)),
// 	}
// }
